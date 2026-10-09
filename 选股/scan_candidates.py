# -*- coding: utf-8 -*-
"""
A股每日备选股票池扫描脚本
方法来源：Minervini《股票魔法师》SEPA趋势模板 + 欧奈尔《笑傲股市》CANSLIM + 达瓦斯《我如何从股市赚了100万》箱体理论

初选条件（硬性，全部满足才入选）：
  1. 当日涨幅 > 6%
  2. 当日最高价创 20 日新高（含当日）
  3. 周线趋势不在下跌：周收盘 > 周MA30（约半年线），且周MA30 本周值 >= 上周值
  4. 月线趋势不在下跌：月收盘 > 月MA10，且月MA10 本月值 >= 上月值
  5. 剔除 ST / *ST / 退市股；仅沪深两市（主板+创业板+科创板）

数据源（按用户要求，弃用东方财富）：
  1) 新浪财经：Market_Center.getHQNodeData（全市场涨幅榜）+ CN_MarketDataService.getKLineData（日K）
  2) 兜底：腾讯财经 ifzq.gtimg.cn 前复权日K
输出：I:/tradebuddy/选股/备选清单/<日期>_A股备选股票清单.md
"""
import urllib.request
import json
import re
import datetime
import os
import time
import sys
from concurrent.futures import ThreadPoolExecutor, as_completed

OUT_DIR = r"I:\tradebuddy\选股\备选清单"
THRESHOLD_PCT = 6.0

HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "Referer": "https://finance.sina.com.cn/",
}

SINA_LIST_URL = ("https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/"
                 "Market_Center.getHQNodeData?page={page}&num=100&sort=changepercent&asc=0"
                 "&node=hs_a&symbol=&_s_r_a=init")
SINA_KLINE_URL = ("https://quotes.sina.cn/cn/api/jsonp_v2.php/var%20d=/"
                  "CN_MarketDataService.getKLineData?symbol={sym}&scale=240&ma=no&datalen=620")
TENCENT_KLINE_URL = ("https://web.ifzq.gtimg.cn/appstock/app/fqkline/get?"
                     "param={sym},day,,,{n},qfq")


def http_get(url, retries=3, timeout=25):
    last_err = None
    for i in range(retries):
        try:
            req = urllib.request.Request(url, headers=HEADERS)
            with urllib.request.urlopen(req, timeout=timeout) as r:
                return r.read().decode("utf-8", errors="replace")
        except Exception as e:  # noqa: BLE001
            last_err = e
            time.sleep(0.5 + i * 1.0)
    raise RuntimeError("http get failed: %s ... %s" % (url[:90], last_err))


def sina_symbol(code):
    return ("sh" if code[0] in "69" else "sz") + code


def parse_loose_json(text):
    """新浪接口返回不带引号 key 的类 JSON，补引号后解析"""
    text = text.strip()
    if not text or text == "null" or text == "[]":
        return []
    fixed = re.sub(r'([{,]\s*)([A-Za-z_]\w*)\s*:', r'\1"\2":', text)
    return json.loads(fixed)


# ---------------- 第 1 步：新浪全市场涨幅榜 ----------------

def fetch_candidates():
    """按涨幅降序翻页，收集涨幅 > 6% 的沪深A股（剔除ST/退）"""
    out, page = [], 1
    while page <= 40:
        text = http_get(SINA_LIST_URL.format(page=page))
        rows = parse_loose_json(text)
        if not rows:
            break
        stop = False
        for d in rows:
            try:
                pct = float(d.get("changepercent", 0))
            except (TypeError, ValueError):
                continue
            if pct <= THRESHOLD_PCT:
                stop = True
                break
            name = d.get("name", "") or ""
            if ("ST" in name.upper()) or ("退" in name):
                continue
            code = d.get("code", "")
            if not code:
                continue
            # 仅沪深主板(60/00)、科创板(688/689)、创业板(30)；排除北交所(92/43/83/87等)
            if code[:2] not in ("60", "00", "30", "68"):
                continue

            def to_f(v):
                try:
                    return float(v)
                except (TypeError, ValueError):
                    return None

            # 新浪市值单位：万元
            nmc, mktcap = to_f(d.get("nmc")), to_f(d.get("mktcap"))
            out.append({
                "code": code,
                "name": name,
                "pct": pct,
                "price": to_f(d.get("trade")),
                "high": to_f(d.get("high")),
                "turnover_yi": (to_f(d.get("amount")) / 1e8) if to_f(d.get("amount")) else None,
                "float_mv_yi": nmc / 1e4 if nmc else None,
                "total_mv_yi": mktcap / 1e4 if mktcap else None,
                "industry": "-",
            })
        if stop:
            break
        page += 1
        time.sleep(0.3)
    return out


# ---------------- 第 2 步：日K（新浪优先，腾讯前复权兜底） ----------------

def klines_sina(code):
    text = http_get(SINA_KLINE_URL.format(sym=sina_symbol(code))).strip()
    # jsonp 格式：/*<script>...</script>*/\nvar d=([...]);
    m = re.search(r"\[.*\]", text, re.S)
    if not m:
        raise RuntimeError("sina kline bad response")
    arr = json.loads(m.group(0))
    if not isinstance(arr, list) or not arr:
        raise RuntimeError("sina kline empty")
    # 每条: day, open, high, low, close, volume
    return [[k["day"], k["open"], k["close"], k["high"], k["low"]] for k in arr]


def klines_tencent(code):
    data = json.loads(http_get(TENCENT_KLINE_URL.format(sym=sina_symbol(code), n=620)))
    node = data.get("data", {}).get(sina_symbol(code), {})
    arr = node.get("qfqday") or node.get("day") or []
    if not arr:
        raise RuntimeError("tencent kline empty")
    # 每条: [date, open, close, high, low, volume, ...]
    return [[k[0], k[1], k[2], k[3], k[4]] for k in arr]


def klines(code):
    """新浪优先；失败时腾讯前复权兜底"""
    try:
        return klines_sina(code), "sina"
    except Exception:  # noqa: BLE001
        return klines_tencent(code), "tencent"


# ---------------- 指标与条件 ----------------

def ma(vals, n, offset=0):
    seg = vals[len(vals) - n - offset: len(vals) - offset]
    if len(seg) < n:
        return None
    return sum(seg) / n


def resample(daily, period):
    """日线聚合为周线/月线收盘序列（升序）"""
    groups = {}
    for k in daily:
        d = k[0].replace("-", "")
        y, m, dd = int(d[:4]), int(d[4:6]), int(d[6:8])
        if period == "weekly":
            iso = datetime.date(y, m, dd).isocalendar()
            key = (iso[0], iso[1])
        else:
            key = (y, m)
        groups[key] = float(k[2])  # 日期升序，组内最后一天收盘
    return [groups[k] for k in sorted(groups.keys())]


def check_one(c):
    code = c["code"]
    daily, source = klines(code)
    if len(daily) < 300:
        return c, False, None, source
    highs = [float(k[3]) for k in daily]
    closes_d = [float(k[2]) for k in daily]
    today = daily[-1]

    # 条件2：当日最高创 20 日新高（含当日）
    if not (highs[-1] >= max(highs[-20:])):
        return c, False, None, source
    trade_date = today[0]

    # 条件3：周线趋势不下跌
    wcloses = resample(daily, "weekly")
    if len(wcloses) < 31:
        return c, False, None, source
    wma30_now = ma(wcloses, 30)
    wma30_prev = ma(wcloses, 30, 1)
    if not (wcloses[-1] > wma30_now and wma30_now >= wma30_prev):
        return c, False, None, source

    # 条件4：月线趋势不下跌
    mcloses = resample(daily, "monthly")
    if len(mcloses) < 11:
        return c, False, None, source
    mma10_now = ma(mcloses, 10)
    mma10_prev = ma(mcloses, 10, 1)
    if not (mcloses[-1] > mma10_now and mma10_now >= mma10_prev):
        return c, False, None, source

    # 附加：距 52 周最高回撤 %（Minervini：越接近新高越强）
    hi52 = max(highs[-250:])
    dd52 = (hi52 - closes_d[-1]) / hi52 * 100 if hi52 else None
    c["trade_date"] = trade_date
    c["dd52"] = round(dd52, 1) if dd52 is not None else None
    return c, True, None, source


def scan():
    candidates = fetch_candidates()
    results, failed = [], []
    sources = set()
    with ThreadPoolExecutor(max_workers=5) as ex:
        futs = {ex.submit(check_one, c): c for c in candidates}
        for fut in as_completed(futs):
            try:
                c, ok, err, src = fut.result()
                sources.add(src)
                if ok:
                    results.append(c)
                elif err is not None:
                    failed.append({"code": c["code"], "name": c["name"], "err": repr(err)[:60]})
            except Exception as e:  # noqa: BLE001
                cc = futs[fut]
                failed.append({"code": cc["code"], "name": cc["name"], "err": repr(e)[:60]})
    results.sort(key=lambda x: -x["pct"])
    return results, failed, sources


# ---------------- 报告输出 ----------------

def write_report(results, failed, sources, trade_date):
    os.makedirs(OUT_DIR, exist_ok=True)
    path = os.path.join(OUT_DIR, "%s_A股备选股票清单.md" % trade_date)
    src_note = "+".join(sorted(sources)) if sources else "-"
    lines = []
    lines.append("# A股备选股票清单（%s）" % trade_date)
    lines.append("")
    lines.append("> 生成时间：%s ｜ 数据源：新浪财经（日K兜底：腾讯财经前复权）[%s] ｜ 方法：Minervini趋势模板 + 欧奈尔CANSLIM + 达瓦斯箱体（详见《选股方法说明.md》）" % (
        datetime.datetime.now().strftime("%Y-%m-%d %H:%M"), src_note))
    lines.append("")
    lines.append("**入选条件（全部满足）：**")
    lines.append("")
    lines.append("1. 当日涨幅 > 6%")
    lines.append("2. 当日最高价创 20 日新高（含当日）")
    lines.append("3. 周线趋势不在下跌：周收盘 > 周MA30，且周MA30 走平或向上")
    lines.append("4. 月线趋势不在下跌：月收盘 > 月MA10，且月MA10 走平或向上")
    lines.append("5. 剔除 ST / 退市股（仅沪深主板/创业板/科创板）")
    lines.append("")
    lines.append("**当日入选：共 %d 只**" % len(results))
    lines.append("")
    if results:
        lines.append("| 代码 | 名称 | 涨幅% | 收盘 | 当日最高 | 成交额(亿) | 流通市值(亿) | 距52周高点回撤% | 周线 | 月线 |")
        lines.append("|---|---|---|---|---|---|---|---|---|---|")
        for r in results:
            lines.append("| %s | %s | %.2f | %s | %s | %s | %s | %s | 周MA30上 趋势↑ | 月MA10上 趋势↑ |" % (
                r["code"], r["name"], r["pct"],
                r["price"], r["high"],
                ("%.2f" % r["turnover_yi"]) if r["turnover_yi"] else "-",
                ("%.0f" % r["float_mv_yi"]) if r["float_mv_yi"] else "-",
                r["dd52"] if r["dd52"] is not None else "-",
            ))
    else:
        lines.append("（今日无满足全部条件的股票——弱市/调整日属于正常现象，空仓等待也是策略的一部分。）")
    lines.append("")
    lines.append("**使用提醒：**")
    lines.append("")
    lines.append("- 本清单仅为**初选**（自动化第一道筛），不构成买入建议。")
    lines.append("- 入选后需按《选股方法说明》做二次人工确认：VCP收缩形态 / 杯柄或箱体突破点 / 基本面（当季利润增速、年度增速、行业龙头）/ 市场整体方向。")
    lines.append("- 买入后铁律：单笔亏损不超过 7%-8%（Minervini 建议 5%-10% 上限），跌破买点/箱体下沿无条件离场。")
    lines.append("")
    if failed:
        lines.append("**数据获取失败个股（未能完成筛选，请人工补查，共 %d 只）：**" % len(failed))
        lines.append("")
        lines.append("| 代码 | 名称 | 失败原因 |")
        lines.append("|---|---|---|")
        for f in failed:
            lines.append("| %s | %s | %s |" % (f["code"], f["name"], f["err"]))
        lines.append("")
    with open(path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines))
    return path


def main():
    results, failed, sources = scan()
    trade_date = results[0]["trade_date"] if results else datetime.date.today().isoformat()
    path = write_report(results, failed, sources, trade_date)
    print("TRADE_DATE: %s" % trade_date)
    print("CANDIDATES: %d" % len(results))
    print("FETCH_FAILED: %d" % len(failed))
    print("DATA_SOURCE: %s" % ("+".join(sorted(sources)) if sources else "-"))
    for r in results[:50]:
        print("%s %+.2f%%" % (r["code"], r["pct"]))
    print("REPORT: %s" % path)


if __name__ == "__main__":
    main()
