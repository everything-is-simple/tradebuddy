# tradebuddy 项目长期记忆

## 三步流水线蓝图（总控：I:\tradebuddy\docs\TradeBuddy总控Prompt.md）
- 每晚：19:00 强势股初选（已完成，Python MVP）→ 19:30 生命周期识别（斯波朗迪统计框架，未建）→ 20:00 立花义正交易信号（未建）。
- 技术栈：Go + SQLite（modernc.org/sqlite 纯Go免CGO）+ 桌面端 Wails v2 + klinecharts（退级：Go embed HTTP）。门禁 G0-G5 顺序开发，G0=原著数字化（立花口诀+1975-76交易谱24份PDF）。
- 用户方法论铁律：价量原始数据>均线近似；统计分类→概率→验证；尊重原著自发挥<30%；概念正确性/功能完备性/接口清晰性/前后一致性四验收标准；脱离AI可独立操作的文档标准。

## 选股子系统（I:\tradebuddy\选股\）
- 三书合一初选系统：Minervini《股票魔法师》SEPA趋势模板 + 欧奈尔《笑傲股市》CANSLIM + 达瓦斯《我如何从股市赚了100万》箱体理论。
- 每日17:05定时任务（id d439f0d5-e557-4215-9c64-2ab03103f518）运行 scan_candidates.py，输出 备选清单\YYYY-MM-DD_A股备选股票清单.md。
- 入选硬条件：涨幅>6% + 当日high创20日新高 + 周线周MA30不向下 + 月线月MA10不向下；剔除ST/退/北交所；仅沪深主板/创业板/科创板。
- 初选只是第一道筛，二次人工确认：VCP/杯柄形态、距52周高点回撤(≤25%最佳)、CANSLIM基本面(C/A/L)、大盘方向M。止损铁律7-8%。

## 数据源约束（用户明确要求，勿违反）
- **禁用东方财富接口**（用户认为其不愿对外开放）。
- 顺序：新浪财经（榜单 Market_Center.getHQNodeData + 日K quotes.sina.cn jsonp）→ 腾讯财经前复权日K兜底 → 同花顺（无公开API）。
- 技术坑：新浪 hs_a 混入北交所(920xxx)需按代码前缀过滤；新浪返回无引号key类JSON需正则补引号；新浪 money.finance 日K端点已失效(Service not valid)。

## 用户本地书库
- H:\《交易之no.14.year》\：按"年份.(国家)作者书名"组织。Minervini 三部曲正文为整本扫描图(mofa1/2/3)；笑傲股市第4版 txt 为占位文件，正文在 PDF。
