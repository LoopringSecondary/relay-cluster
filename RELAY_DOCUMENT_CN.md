# Loopring Relay 中文文档
介绍一下Loopring Relay，然后介绍下该文档的大概内容

***

## 目录

- [词汇表](Glossary)
- [Relay如何工作]()
  * [系统架构]()
  * [钱包后台]()
  * [交易所后台]()
  * [以太坊解析]()
  * [撮合服务]()
  * [订单广播]()
- [功能介绍]()
  * [订单管理]()
  * [账户管理]()
  * [以太坊transaction管理]()
  * [交易&撮合]()
  * [交易所行情]()
  * [元信息]()
- [怎么接入Relay]()
  * [jsonrpc]()
  * [socketio]()
  * [sdk]()
- [如果部署自己的Relay]()
  * [代码编译]()
  * [配置文件解释]()
- [如何部署Miner]()
  * [代码编译]()
  * [配置文件解释]()
- [Relay Cluster]()
  * [与standalone版本区别]()
  * [系统架构]()
  * [谁更适合部署cluster版本]()
  * [获取技术支持]()
- [获取帮助]()
- [相关资源]()

---

## 词汇表

分类 | 名词 | 解释
---|---|---
订单 | Order | 符合Loopring protocol格式的订单数据
订单 | OrderHash | 订单的签名，即由订单部分字段做执行散列算法后生成的摘要
订单 | Owner | 订单所有者，即用户钱包地址
订单 | OrderType | 订单类型，Relay支持的两种订单订单类型 : market_order( 市场订单)，是整个交易所订单池共享的订单，可以被任何人成交; p2p_order(点对点订单)，是不包含钱包认证私钥的订单点对点订单，只能够被授权共享了钱包认证私钥的用户才能撮合。
订单 | WalletAddress | 提供订单的钱包分润地址，通常是钱包或者交易所产品研发团队的钱包地址，用来参与订单成功撮合后的利润分成，目前方案是，钱包会分取撮合利润的20%，撮合者(Miner)会分取撮合利润的80%。
订单 | AuthAddr & AuthPrivateKey | 提交订单时，随机成功的公私钥对，AuthAddr用来参与订单的签名，AuthPrivateKey用来参与提交撮合时环路的签名，目的是为了防止订单或者环路被篡改，同时在点对点订单的场景，AuthPrivateKey在通过二维码只分享给特定用户的情况下，可以保护订单只被单独用户吃单。
订单 | TokenS | 要出售的Token, 请参考支持的Token列表
订单 | TokenB | 要买入的Token，请参考支持的Token列表
订单 | AmountS | 要出售的Token数量
订单 | AmountB | 要买入的Token数量
订单 | ValidSince | 订单生效开始时间，表示单位为时间戳，如果当前时间小于ValidSince，订单是未生效状态。
订单 | ValidUntil | 订单有效截止时间，表示单位为时间戳，超过后订单自动失效。
订单 | LrcFee | 设置该笔订单撮合需要的LrcFee
订单 | buyNoMoreThanAmountB | 是不是允许购买超过amountB数量的tokeB，比如当前市场卖价(LRC-WETH)是0.001，用户下单价格是0.002买入100个（需要0.2个WETH），如果buyNoMoreThanAmountB=true，那最终用户会以0.001的价格（不考虑撮合收益）购买到100个LRC，消耗0.1个WETH；如果buyNoMoreThanAmountB=false，那最终用户会消耗掉所有的WETH（0.2个）以0.001的价格（不考虑撮合收益）购买到200个LRC。
订单 | marginSplitPercentage | 撮合分润中用来支付撮合费的比例，通常默认是50%。
订单 | v, r, s | 订单签名的结果，是首先采用Keccak256算法对订单部分字段生成OrderHash, 再针对Hash做ECDSA签名，生成的结果。
订单 | powNonce | 订单提交工作量证明，为了防止订单子系统被spam，我们采用工作量证明的方式来限制过多的订单提交，powNonce参与工作量证明算法计算，订单通过工作量证明校验后，提交到Relay，我们会以相同的工作量证明算法来校验nonce是否通过了工作量证明。
订单 | 撮合 | 即两个以上订单满足形成Loopring环路的条件，可以形成环形成交，环形成交即Loopring的撮合。
订单 | 环路 | 相对于传统交易所订单两两互相成交，Loopring可以针对多笔订单串联形成环形成交队列，这个环形头尾相连的订单队列，即环路。
订单 | 软取消 | 在Loopr2版本的钱包中，用户取消订单只能提交到智能合约，不仅花费油费，而且不能即时取消。所以我们在Relay中增加了软取消功能，在满足软取消情况(比如订单未撮合或者未在撮合流程中)下，可以通过Relay取消订单，不消耗油费并且即时取消订单。
账户 | Allowance | 代币授权，这里通常指的是用户授权给Loopring protocol，想要Loopring智能合约能够撮合用户订单，只有用户针对合约做授权操作后，Loopring智能合约才能够撮合用户订单。
账户 | Balance | 用户资产余额，包含ETH余额和所有ERC20 Token余额。
账户 | WETH | WETH是以太坊上锚定ETH的ERC20 Token，在没有任何额外费(除了一笔transaction油费)用情况下，永远可以和ETH等量交换，Loopring合约只支持ERC20 Token之间资产交换，并不支持ETH和ERC20 Token交换，所以在用户交易前，需要将ETH转换成WETH，同时授权Loopring合约使用WETH Token.
市场 | Fill | 成交信息，环路撮合后，智能合约发出的成交数据Event
市场 | Depth | 市场深度
市场 | Ticker | 24小时市场变化统计数据
市场 | Trend | 市场变化趋势信息，目前支持最小维度1Hr
市场 | RingMined | Order组成的环路撮合的结果
市场 | Cutoff | 用户以地址为单位设置的订单全部失效时间点，cutoff时间点之前的订单，全部变为无效订单
市场 | PriceQuote | 各个币种的市价参考，目前支持BTC, CNY, USD
合约 | LoopringProtocolImpl | Loopring入口合约地址，伴随着合约升级，地址会有变化
合约 | DelegateAddress | Delegate合约地址，订单池按照Delegate合约划分，不同Delegate地址的订单之间，不能互相撮合。
通用 | Token | 即以太坊上代币，目前只支持完全符合ERC20标准的Token
通用 | Transaction | 通常是指用户转账/授权/合约调用等以太坊交易操作，在Relay，我们对transaction进行了包装，包含用户订单撮合类型，同时包含所有以太坊交易操作类型，方便用户区分。
通用 | Gas | 提交以太坊交易，需要指定GasPrice和GasLimit用来支付交易产生的费用，Loopring支持获取当前网络最佳GasPrice
通用 | Nonce | 以用户钱包地址为单位，从1开始递增的整数，当前值等于用户提交成功的transaction总数，用户提交transaction需要提供nonce做校验，同一个nonce只能有一个transaction提交成功，由于Relay接入了Loopring众多的钱包版本（web/ios/android）, 同时有多个合作伙伴接入，所以提供了集中维护Nonce的功能，最大程度的提高transaction成功率。
通用 | Miner | Loopring撮合服务，在订单池中发现环路，并提交到智能合约撮合。
通用 | Decimal | ERC20 Token 单位精度，一般情况下订单里Token数量除以Decimal, 是实际数量，通常Decimal=1e18。
通用 | Symbol | ERC20 Token简称，例如Loopring ERC20 Token, 是LRC。