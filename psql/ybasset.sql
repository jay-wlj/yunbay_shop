/**
CREATE USER ybasset WITH PASSWORD '123456';
CREATE DATABASE ybasset with owner=ybasset ENCODING='UTF8';
GRANT ALL PRIVILEGES ON DATABASE ybasset to ybasset;
ALTER DATABASE ybapi SET TIMEZONE='PRC';
\c ybasset ybasset;
**/


-- 创建array连接聚合
DROP AGGREGATE IF EXISTS anyarray_agg(anyarray);
CREATE AGGREGATE anyarray_agg(anyarray) (
  SFUNC = array_cat,
  STYPE = anyarray
);

-- CREATE FUNCTION unix_timestamp() RETURNS integer AS $$ 
-- SELECT (date_part('epoch',now()))::integer;   
-- $$ LANGUAGE SQL IMMUTABLE;

CREATE FUNCTION unix_timestamp() RETURNS integer AS $$ 
SELECT extract(epoch from now())::integer;   
$$ LANGUAGE SQL IMMUTABLE;

CREATE FUNCTION from_unixtime(int) RETURNS timestamp AS $$ 
SELECT to_timestamp($1)::timestamp; 
$$ LANGUAGE SQL IMMUTABLE;

CREATE TABLE "user_lock"
(
  id bigserial,
  user_id bigint,   
  status smallint DEFAULT 0,   
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX user_lock_idx ON "user_lock"(user_id, status);
ALTER TABLE user_lock REPLICA IDENTITY FULL;
COMMENT ON TABLE user_lock IS '用户冻结表';
COMMENT ON COLUMN user_lock.user_id IS '用户id';
COMMENT ON COLUMN user_lock.status IS '帐户冻结类型(0:未冻结 1:已冻结)';
COMMENT ON COLUMN user_lock.create_time IS '注册时间';
COMMENT ON COLUMN user_lock.update_time IS '更新时间';

CREATE TABLE "user_asset"
(
  id bigserial,
  user_id bigint,  
  total_ybt NUMERIC default 0,
  normal_ybt NUMERIC default 0,
  lock_ybt numeric default 0,  
  freeze_ybt numeric default 0,
  total_kt numeric default 0,
  normal_kt numeric default 0,
  lock_kt numeric default 0,    
  total_snet numeric DEFAULT 0,
  normal_snet numeric DEFAULT 0,
  lock_snet numeric DEFAULT 0,
  status smallint DEFAULT 0,    
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX user_asset_user_idx ON "user_asset"(user_id);
ALTER TABLE user_asset REPLICA IDENTITY FULL;
COMMENT ON TABLE user_asset IS '用户资产表';
COMMENT ON COLUMN user_asset.user_id IS '用户id';
COMMENT ON COLUMN user_asset.total_ybt IS '持有ybt总量';
COMMENT ON COLUMN user_asset.total_kt IS '持有kt总量';
COMMENT ON COLUMN user_asset.lock_ybt IS '锁定(冻结)ybt总量';
COMMENT ON COLUMN user_asset.lock_kt IS '锁定(冻结)kt总量';
COMMENT ON COLUMN user_asset.freeze_ybt IS '空投锁定的ybt总量';
COMMENT ON COLUMN user_asset.normal_ybt IS '可用ybt总量';
COMMENT ON COLUMN user_asset.normal_kt IS '可用kt总量';
COMMENT ON COLUMN user_asset.total_snet IS '持有的snet';
COMMENT ON COLUMN user_asset.normal_snet IS '可用的snet';
COMMENT ON COLUMN user_asset.lock_snet IS '锁定的snet';
COMMENT ON COLUMN user_asset.status IS '帐户冻结类型(0:未冻结 1:已冻结)';
COMMENT ON COLUMN user_asset.create_time IS '注册时间';
COMMENT ON COLUMN user_asset.update_time IS '更新时间';




CREATE TABLE "asset_lock"
(
  id bigserial,
  user_id bigint, 
  "type" SMALLINT DEFAULT 0,  
  lock_type smallint default 0,  
  lock_amount numeric not null,
  unlock_time int default 0,
  date varchar(20) default '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX asset_lock_idx ON "asset_lock"(user_id, type, lock_type, create_time);
ALTER TABLE asset_lock REPLICA IDENTITY FULL;
COMMENT ON TABLE asset_lock IS '用户资产冻结明细表';
COMMENT ON COLUMN asset_lock.user_id IS '用户id';
COMMENT ON COLUMN asset_lock.type IS '币种类型(0:ybt 1:kt)';
COMMENT ON COLUMN asset_lock.lock_amount IS '冻结数量';
COMMENT ON COLUMN asset_lock.lock_type IS '冻结类型(0:空投奖励锁定(解锁) 1:定期冻结(解冻) 2:永久冻结 3:提币冻结(解冻))';
COMMENT ON COLUMN asset_lock.unlock_time IS '解锁时间(定期锁定)';
COMMENT ON COLUMN asset_lock.date IS '日期';
COMMENT ON COLUMN asset_lock.create_time IS '注册时间';
COMMENT ON COLUMN asset_lock.update_time IS '更新时间';

CREATE TABLE "user_asset_detail"
(
  id bigserial,
  user_id bigint, 
  "type" SMALLINT DEFAULT 0,
  "transaction_type" SMALLINT DEFAULT 1,
  amount numeric not null,
  date varchar(20) DEFAULT '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);

CREATE INDEX user_asset_detail_idx ON "user_asset_detail"(type, transaction_type, user_id, date, create_time);
ALTER TABLE user_asset_detail REPLICA IDENTITY FULL;
COMMENT ON TABLE user_asset_detail IS '用户资产明细表';
COMMENT ON COLUMN user_asset_detail.user_id IS '用户id';
COMMENT ON COLUMN user_asset_detail.type IS '币种类型(0:ybt 1:kt)';
COMMENT ON COLUMN user_asset_detail.transaction_type IS 'ybt交易类型(0:充币 1:提币 2:消费(挖矿)奖励 3:商家奖励 4:邀请奖励 5:活动奖励) kt交易类型(0:充币 1:提币 2:收益金 3:商品消费 4:商品卖出 5:退款)';
COMMENT ON COLUMN user_asset_detail.amount IS '交易数量';
COMMENT ON COLUMN user_asset_detail.create_time IS '注册时间';
COMMENT ON COLUMN user_asset_detail.update_time IS '更新时间';

CREATE TABLE "bonus_kt_detail"
(
  id bigserial,
  user_id bigint, 
  ybt double precision default 0,
  kt double precision default 0,
  date varchar(20) DEFAULT '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX bonus_kt_detail_idx ON "bonus_kt_detail"(user_id, date);
ALTER TABLE bonus_kt_detail REPLICA IDENTITY FULL;
COMMENT ON TABLE bonus_kt_detail IS '用户每天kt收益明细表';
COMMENT ON COLUMN bonus_kt_detail.user_id IS '用户id';
COMMENT ON COLUMN bonus_kt_detail.ybt IS '可分红ybt';
COMMENT ON COLUMN bonus_kt_detail.kt IS 'kt收益金';
COMMENT ON COLUMN bonus_kt_detail.date IS '日期';
COMMENT ON COLUMN bonus_kt_detail.create_time IS '注册时间';
COMMENT ON COLUMN bonus_kt_detail.update_time IS '更新时间';

CREATE TABLE "bonus_ybt_detail"
(
  id bigserial,
  user_id bigint, 
  infos jsonb default null,
  total_ybt  double precision default 0,
  date varchar(20) DEFAULT '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX bonus_ybt_detail_uidx ON "bonus_ybt_detail"(user_id, date);
ALTER TABLE bonus_ybt_detail REPLICA IDENTITY FULL;
COMMENT ON TABLE bonus_ybt_detail IS '用户每天ybt奖励明细表';
COMMENT ON COLUMN bonus_ybt_detail.user_id IS '用户id';
COMMENT ON COLUMN bonus_ybt_detail.infos IS '类型信息';
COMMENT ON COLUMN bonus_ybt_detail.total_ybt IS '累积奖励ybt';
COMMENT ON COLUMN bonus_ybt_detail.date IS '日期';
COMMENT ON COLUMN bonus_ybt_detail.create_time IS '注册时间';
COMMENT ON COLUMN bonus_ybt_detail.update_time IS '更新时间';

CREATE TABLE "yunbay_asset_detail"
(
  id bigserial,  
  amount numeric not null,
  profit numeric not null,
  issue_ybt numeric default 0,
  destoryed_ybt numeric default 0,  
  bonus_ybt numeric default 0,    
  lock_ybt numeric default 0,  
  freeze_ybt numeric default 0,  
  perynbay numeric default 0,  
  difficult numeric default 0,
  period  int not null,
  mining numeric default 0,
  air_drop numeric default 0,
  air_unlock numeric default 0,
  air_recover numeric default 0,
  activity numeric default 0,
  project numeric default 0,
  miners int default 0,
  bonusers int default 0,
  date varchar(20) not null,
  kt_status smallint default 0,
  ybt_status smallint default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX yunbay_asset_detail_createtime_idx ON yunbay_asset_detail(kt_status, ybt_status, period, create_time);
CREATE UNIQUE INDEX yunbay_asset_detail_date_idx ON yunbay_asset_detail(date);
ALTER TABLE yunbay_asset_detail REPLICA IDENTITY FULL;
COMMENT ON TABLE yunbay_asset_detail IS '平台ybt挖矿明细表';
COMMENT ON COLUMN yunbay_asset_detail.amount IS '平台当日交易额(kt)';
COMMENT ON COLUMN yunbay_asset_detail.profit IS '平台当日净利额(kt)';
COMMENT ON COLUMN yunbay_asset_detail.issue_ybt IS '平台当日发行的ybt';
COMMENT ON COLUMN yunbay_asset_detail.destoryed_ybt IS '平台当日销毁的ybt';
COMMENT ON COLUMN yunbay_asset_detail.lock_ybt IS '锁定的ybt';
COMMENT ON COLUMN yunbay_asset_detail.freeze_ybt IS '空投冻结的ybt';
COMMENT ON COLUMN yunbay_asset_detail.date IS '日期';
COMMENT ON COLUMN yunbay_asset_detail.bonus_ybt IS '可分红云贝数';
COMMENT ON COLUMN yunbay_asset_detail.perynbay IS '每个云贝可得分红,如(1YBT=0.0001KT)';
COMMENT ON COLUMN yunbay_asset_detail.difficult IS '挖矿难度系数';
COMMENT ON COLUMN yunbay_asset_detail.period IS '周期';
COMMENT ON COLUMN yunbay_asset_detail.mining IS '挖矿释放';
COMMENT ON COLUMN yunbay_asset_detail.air_drop IS '空投奖励';
COMMENT ON COLUMN yunbay_asset_detail.air_unlock IS '空投释放';
COMMENT ON COLUMN yunbay_asset_detail.air_recover IS '空投回收';
COMMENT ON COLUMN yunbay_asset_detail.activity IS '活动释放';
COMMENT ON COLUMN yunbay_asset_detail.project IS '项目释放';
COMMENT ON COLUMN yunbay_asset_detail.miners IS '挖矿人数';
COMMENT ON COLUMN yunbay_asset_detail.bonusers IS '分红人数';
COMMENT ON COLUMN yunbay_asset_detail.kt_status IS 'kt分红发放状态(0:未发放 1:已发放)';
COMMENT ON COLUMN yunbay_asset_detail.ybt_status IS 'ybt发放状态(0:未发放 1:已发放)';
COMMENT ON COLUMN yunbay_asset_detail.create_time IS '注册时间';
COMMENT ON COLUMN yunbay_asset_detail.update_time IS '更新时间';

CREATE TABLE "yunbay_asset"
(
  id bigserial,
  total_kt numeric not null,
  total_kt_profit numeric not null,
  total_issue_ybt numeric not null,
  total_destroyed_ybt numeric default 0,
  total_mining numeric default 0,
  total_air_drop numeric default 0,
  total_air_unlock numeric default 0,
  total_air_recover numeric default 0,
  total_activity numeric default 0,
  total_project numeric default 0,
  total_perynbay DOUBLE PRECISION default 0,
  date varchar(20) not null,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX yunbay_asset_date_idx ON yunbay_asset(date);
ALTER TABLE yunbay_asset REPLICA IDENTITY FULL;
COMMENT ON TABLE yunbay_asset IS '平台资产表';
COMMENT ON COLUMN yunbay_asset.total_kt IS '平台总交易额';
COMMENT ON COLUMN yunbay_asset.total_kt_profit IS '平台总利润';
COMMENT ON COLUMN yunbay_asset.total_issue_ybt IS '平台总发行云贝数';
COMMENT ON COLUMN yunbay_asset.total_destroyed_ybt IS '平台总销毁云贝数';
COMMENT ON COLUMN yunbay_asset.total_mining IS '平台总挖矿ybt';
COMMENT ON COLUMN yunbay_asset.total_air_drop IS '平台总空投ybt';
COMMENT ON COLUMN yunbay_asset.total_air_unlock IS '平台总空投释放ybt';
COMMENT ON COLUMN yunbay_asset.total_air_recover IS '平台总空投回收ybt';
COMMENT ON COLUMN yunbay_asset.total_activity IS '平台总活动释放ybt';
COMMENT ON COLUMN yunbay_asset.total_project IS '平台总项目方释放ybt';
COMMENT ON COLUMN yunbay_asset.total_perynbay IS '累积1ybt收益';
COMMENT ON COLUMN yunbay_asset.date IS '日期';
COMMENT ON COLUMN yunbay_asset.create_time IS '注册时间';
COMMENT ON COLUMN yunbay_asset.update_time IS '更新时间';

CREATE TABLE "yunbay_asset_pool"
(
  id bigserial,
  order_id bigint not null,
  payer_userid bigint not null,
  currency_type smallint DEFAULT 1,
  pay_amount numeric not null,
  seller_userid bigint DEFAULT 0,
  seller_amount numeric not null,  
  rebat_amount numeric not null,
  seller_kt double PRECISION default 0,
  status SMALLINT DEFAULT 0,
  date varchar(20) not null,
  country smallint default 0,
  extinfos jsonb default '{}',
  publish_area smallint default 1,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX asset_pool_s ON yunbay_asset_pool(order_id);
CREATE INDEX asset_pool_s_d ON yunbay_asset_pool(payer_userid, seller_userid, date);
ALTER TABLE yunbay_asset REPLICA IDENTITY FULL;
COMMENT ON TABLE yunbay_asset_pool IS '平台交易资金池明细';
COMMENT ON COLUMN yunbay_asset_pool.order_id IS '订单id';
COMMENT ON COLUMN yunbay_asset_pool.payer_userid IS '买家用户ID';
COMMENT ON COLUMN yunbay_asset_pool.currency_type IS '货币类型(0:ybt 1:kt)';
COMMENT ON COLUMN yunbay_asset_pool.pay_amount IS '订单支付金额';
COMMENT ON COLUMN yunbay_asset_pool.seller_userid IS '卖家用户ID(0为平台)';
COMMENT ON COLUMN yunbay_asset_pool.seller_amount IS '卖家应所得额';
COMMENT ON COLUMN yunbay_asset_pool.rebat_amount IS '贡献值';
COMMENT ON COLUMN yunbay_asset_pool.country IS '国家(0:国际版 1:国内版)';
COMMENT ON COLUMN yunbay_asset_pool.extinfos IS '扩展信息';
COMMENT ON COLUMN yunbay_asset_pool.publish_area IS '发布专区(默认KT专区)';
COMMENT ON COLUMN yunbay_asset_pool.status IS '当前状态(0:平台冻结状态 1:已打款给卖家 2:已返回给买家(取消订单等)';
COMMENT ON COLUMN yunbay_asset_pool.date IS '日期';
COMMENT ON COLUMN yunbay_asset_pool.create_time IS '注册时间';
COMMENT ON COLUMN yunbay_asset_pool.update_time IS '更新时间';
comment on column yunbay_asset_pool.seller_kt is 'ybt支付的订单待返商家的kt';

CREATE TABLE "ordereward"
(
  id bigserial,
  order_id bigint not null,
  ybt numeric not null,
  buyer_userid bigint not null,
  buyer_ybt numeric not null,
  seller_userid numeric not null,
  seller_ybt numeric not null,
  recommender_userid bigint default 0,
  recommender_ybt numeric default 0,
  recommender2_userid bigint default 0,
  recommender2_ybt numeric default 0,
  yunbay_userid bigint default 0,
  yunbay_ybt numeric default 0,
  seller_status smallint default 0,
  valid smallint default 0,
  date varchar(20) DEFAULT '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX ordereward_orderid_idx ON ordereward(order_id);
CREATE INDEX ordereward_buyer_seller_date_idx ON ordereward(buyer_userid, seller_userid, date);
ALTER TABLE ordereward REPLICA IDENTITY FULL;
COMMENT ON TABLE ordereward IS '订单挖矿表';
COMMENT ON COLUMN ordereward.order_id IS '订单id';
COMMENT ON COLUMN ordereward.ybt IS '订单所挖的总ybt';
COMMENT ON COLUMN ordereward.buyer_userid IS '买家用户id';
COMMENT ON COLUMN ordereward.buyer_ybt IS '买家所得ybt';
COMMENT ON COLUMN ordereward.seller_userid IS '卖家用户id';
COMMENT ON COLUMN ordereward.seller_ybt IS '卖家所得ybt';
COMMENT ON COLUMN ordereward.recommender_userid IS '一级推荐人用户id';
COMMENT ON COLUMN ordereward.recommender_ybt IS '一级推荐人所得ybt';
COMMENT ON COLUMN ordereward.recommender2_userid IS '二级推荐人用户id';
COMMENT ON COLUMN ordereward.recommender2_ybt IS '二级推荐人所得ybt';
COMMENT ON COLUMN ordereward.yunbay_userid IS '平台用户ID';
COMMENT ON COLUMN ordereward.yunbay_ybt IS '平台所得ybt';
COMMENT ON COLUMN ordereward.seller_status IS '商家奖励转给商家状态(0:未转,保证金帐号持有 1:已从保证金帐户转给商家用户);';
COMMENT ON COLUMN ordereward.valid IS '订单是否合法(0:商家奖励归保证金帐户所有 1:合法的订单 商家奖励及对应的kt收益会转给商家);';
COMMENT ON COLUMN ordereward.date IS '日期';
COMMENT ON COLUMN ordereward.create_time IS '注册时间';
COMMENT ON COLUMN ordereward.update_time IS '更新时间';





create table "tradeflow"
(
  id bigserial,  
  total_orders int default 0,
  payed_orders int default 0,  
  total_payers int default 0,
  total_amount numeric default 0,
  total_profit numeric default 0,
  perynbay numeric default 0,
  country smallint default 0,
  date varchar(20) default '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX tradeflow_idx ON tradeflow(date, country);
ALTER TABLE tradeflow REPLICA IDENTITY FULL;
COMMENT ON TABLE tradeflow IS '交易流水';
COMMENT ON COLUMN tradeflow.id IS 'id';
COMMENT ON COLUMN tradeflow.total_orders IS '总订单数(待支付及已支付订单)';
COMMENT ON COLUMN tradeflow.payed_orders IS '已支付订单数';
COMMENT ON COLUMN tradeflow.total_payers IS '支付人数';
COMMENT ON COLUMN tradeflow.total_amount IS '总交易额(kt)';
COMMENT ON COLUMN tradeflow.total_profit IS '总利润(kt)';
COMMENT ON COLUMN tradeflow.country IS '国家(0:国际版 1:国内版)';
COMMENT ON COLUMN tradeflow.date IS '日期';
COMMENT ON COLUMN tradeflow.create_time IS '注册时间';
COMMENT ON COLUMN tradeflow.update_time IS '更新时间';

-- 建立总交易流水的view
create view tradeflow_all as (select a.update_time, a.perynbay, b.* from tradeflow a join (select  date,sum(total_orders) as total_orders, sum(payed_orders) as payed_orders, sum(total_payers) as total_payers, sum(total_amount) as total_amount, sum(total_profit) as total_profit from tradeflow  group by date) b on a.date=b.date where country=(select country from tradeflow c where c.date=b.date limit 1));


create table "user_wallet"
(
  id bigserial,
  "type" smallint default 1,
  user_id bigint default 0,  
  bind_address varchar(64) default '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX user_wallet_idx ON user_wallet(type, user_id);
ALTER TABLE user_wallet REPLICA IDENTITY FULL;
COMMENT ON TABLE user_wallet IS 'ybt发行量规则';
COMMENT ON COLUMN user_wallet.id IS 'id';
COMMENT ON COLUMN user_wallet.type IS '币种类型(0:ybt 1:kt)';
COMMENT ON COLUMN user_wallet.user_id IS '用户id';
COMMENT ON COLUMN user_wallet.bind_address IS '绑定的钱包地址';
COMMENT ON COLUMN user_wallet.create_time IS '注册时间';
COMMENT ON COLUMN user_wallet.update_time IS '更新时间';

create table "wallet_address"
(
  id bigserial,  
  "type" smallint default 0,
  user_id bigint default 0,    
  name varchar(64) default '',
  adddress varchar(64) default '',
  "default" boolean default false,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX wallet_address_idx ON wallet_address(user_id);
ALTER TABLE wallet_address REPLICA IDENTITY FULL;
COMMENT ON TABLE wallet_address IS 'ybt发行量规则';
COMMENT ON COLUMN wallet_address.id IS 'id';
COMMENT ON COLUMN wallet_address.type IS '币种类型(0:ybt 1:kt)';
COMMENT ON COLUMN wallet_address.user_id IS '用户id';
COMMENT ON COLUMN wallet_address.name IS '缩略名称';
COMMENT ON COLUMN wallet_address.adddress IS '绑定的钱包地址';
COMMENT ON COLUMN wallet_address.default IS '是否默认钱包地址';
COMMENT ON COLUMN wallet_address.create_time IS '注册时间';
COMMENT ON COLUMN wallet_address.update_time IS '更新时间';


create table "withdraw_flow"
(
  id bigserial,
  channel smallint default 0,
  flow_type smallint default 0,
  user_id bigint default 0,
  to_user_id bigint default -1,
  lock_asset_id bigint default 0,
  tx_type smallint default 0,  
  from_address varchar(64) default '',
  address varchar(64) default '',
  amount numeric default 0,
  fee numeric default 0,
  feeinether numeric default 0,
  txhash varchar(128) default '',  
  status smallint default 0,  
  reason varchar(256) default '',
  date varchar(20) default '',
  maner varchar(50) default '',
  country smallint default 0,
  extinfos jsonb not null default '{}',
  check_time int default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX withdraw_flow_idx ON withdraw_flow(user_id, tx_type, status);
CREATE INDEX withdraw_country_idx ON withdraw_flow(channel, date, country);
ALTER TABLE withdraw_flow REPLICA IDENTITY FULL;
COMMENT ON TABLE withdraw_flow IS '提币交易流水';
COMMENT ON COLUMN withdraw_flow.id IS '交易id';
COMMENT ON COLUMN withdraw_flow.channel IS '充提渠道(0:官方 1:热币)';
COMMENT ON COLUMN withdraw_flow.flow_type IS '提币方式(0:走合约 1:走内盘)';
COMMENT ON COLUMN withdraw_flow.user_id IS '交易用户id';
COMMENT ON COLUMN withdraw_flow.lock_asset_id IS '交易冻结的用户资产id(asset_lock id)';
COMMENT ON COLUMN withdraw_flow.tx_type IS '提币类型(0:ybt 1:kt)';
COMMENT ON COLUMN withdraw_flow.from_address IS '提币原地址';
COMMENT ON COLUMN withdraw_flow.address IS '提币地址';
COMMENT ON COLUMN withdraw_flow.amount IS '提币数量';
COMMENT ON COLUMN withdraw_flow.fee IS '手续费';
COMMENT ON COLUMN withdraw_flow.feeinether IS '交易使用的以太币';
COMMENT ON COLUMN withdraw_flow.txhash IS '交易id';
COMMENT ON COLUMN withdraw_flow.status IS '交易状态(-1:审核不通过 0:未审核 1:审核通过 2:等待提交 3:区块交易已提交 4:区块交易确认中 5:区块交易失败 6:区块交易成功)';
COMMENT ON COLUMN withdraw_flow.reason IS '交易失败情况错误原因';
COMMENT ON COLUMN withdraw_flow.date IS '申请日期';
COMMENT ON COLUMN withdraw_flow.maner IS '审核人';
COMMENT ON COLUMN withdraw_flow.check_time IS '审核时间';
COMMENT ON COLUMN withdraw_flow.country IS '国家(0:国际版 1:国内版)';
COMMENT ON COLUMN withdraw_flow.create_time IS '注册时间';
COMMENT ON COLUMN withdraw_flow.update_time IS '更新时间';
COMMENT ON COLUMN withdraw_flow.extinfos is '提币扩展字段';
COMMENT ON COLUMN withdraw_flow.to_user_id is '转入帐户id(yunbay帐户转入才有效)';
select setval('withdraw_flow_id_seq', 16025658);   --设置订单id初始值




create table "recharge_flow"
(
  id bigserial,  
  channel smallint default 0,
  flow_type smallint default 0,
  user_id bigint default 0,
  asset_id bigint default 0,
  txhash varchar(128) default '',
  from_address varchar(64) default '',  
  address varchar(64) default '',  
  "tx_type" smallint default 0,
  amount numeric default 0,
  country smallint default 0,
  date varchar(20) default '',
  block_time varchar(20) default '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX uidx_recharge_flow ON recharge_flow(txhash);
CREATE INDEX recharge_flow_idx ON recharge_flow(user_id, tx_type, create_time);
CREATE INDEX recharge_country_idx ON recharge_flow(channel, date, country);
ALTER TABLE recharge_flow REPLICA IDENTITY FULL;
COMMENT ON TABLE recharge_flow IS '充值流水';
COMMENT ON COLUMN recharge_flow.id IS '充值id';
COMMENT ON COLUMN recharge_flow.channel IS '充提渠道(0:官方 1:热币)';
COMMENT ON COLUMN recharge_flow.flow_type IS '充币方式(0:走合约 1:走内盘)';
COMMENT ON COLUMN recharge_flow.user_id IS '充值用户id';
COMMENT ON COLUMN recharge_flow.asset_id IS '关联的user_asset_detail_id';
COMMENT ON COLUMN recharge_flow.tx_type IS '充值类型(0:ybt 1:kt)';
COMMENT ON COLUMN recharge_flow.from_address IS '充值源地址';
COMMENT ON COLUMN recharge_flow.address IS '充值地址';
COMMENT ON COLUMN recharge_flow.amount IS '充值数量';
COMMENT ON COLUMN recharge_flow.txhash IS '交易id';
COMMENT ON COLUMN recharge_flow.block_time IS '区块确认时间';
COMMENT ON COLUMN recharge_flow.date IS '充值日期';
COMMENT ON COLUMN recharge_flow.create_time IS '注册时间';
COMMENT ON COLUMN recharge_flow.update_time IS '更新时间';
COMMENT ON COLUMN recharge_flow.country IS '国家(0:国际版 1:国内版)';


CREATE TABLE "third_bonus"
(
  id bigserial,
  tid int default 0,
  uid bigint DEFAULT 0,
  ybt numeric default 0,
  kt numeric default 0,
  status smallint default 0,
  date varchar(20) DEFAULT '',  
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX third_bonus_idx ON "third_bonus"(tid, uid, date);
ALTER TABLE third_bonus REPLICA IDENTITY FULL;
COMMENT ON TABLE third_bonus IS '第三方平台分红明细表';
COMMENT ON COLUMN third_bonus.tid IS '平台id,0:yunex';
COMMENT ON COLUMN third_bonus.uid IS '平台用户id';
COMMENT ON COLUMN third_bonus.ybt IS '可分红ybt';
COMMENT ON COLUMN third_bonus.kt IS 'kt收益金';
COMMENT ON COLUMN third_bonus.status IS '发放状态(0:未发放 1:已发放 -1:发放失败(第三方返回))';
COMMENT ON COLUMN third_bonus.date IS '日期';
COMMENT ON COLUMN third_bonus.create_time IS '注册时间';
COMMENT ON COLUMN third_bonus.update_time IS '更新时间';


create table "address_source"
(
  id bigserial,  
  address varchar(64) default '',
  channel int default -1,   
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX address_source_idx ON address_source(address);
ALTER TABLE address_source REPLICA IDENTITY FULL;
COMMENT ON TABLE address_source IS '钱包地址来源';
COMMENT ON COLUMN address_source.id IS 'id';
COMMENT ON COLUMN address_source.address IS '钱包地址';
COMMENT ON COLUMN address_source.channel IS '地址所属渠道平台(-1,其它 0,官方地址 1,热币地址 2,yunex地址)';
COMMENT ON COLUMN address_source.create_time IS '注册时间';
COMMENT ON COLUMN address_source.update_time IS '更新时间';



create table "rmb_recharge"
(
  id bigserial,  
  channel smallint default 0,  
  user_id bigint default 0,
  order_ids bigint[] default NULL,
  subject varchar(256) default '',
  asset_id bigint default 0,
  tx_type smallint default 0,
  txhash varchar(128) default '',
  account varchar(64) default '',    
  amount double precision default 0,
  date varchar(20) default '',  
  status smallint default 0,
  reason varchar(256) default '',
  --order_status smallint default 0,
  over_time int default 0,
  extinfos jsonb not null default '{}',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX rmb_recharge_uidx ON rmb_recharge(user_id, asset_id, txhash, channel);
CREATE INDEX rmb_recharge_idx ON rmb_recharge(tx_type, status, date, over_time);
ALTER TABLE rmb_recharge REPLICA IDENTITY FULL;
COMMENT ON TABLE rmb_recharge IS '充值流水';
COMMENT ON COLUMN rmb_recharge.id IS '充值id';
COMMENT ON COLUMN rmb_recharge.channel IS '充提渠道(10:支付宝 11:微信)';
COMMENT ON COLUMN rmb_recharge.user_id IS '充值用户id';
COMMENT ON COLUMN rmb_recharge.order_ids IS '关联的订单id';
COMMENT ON COLUMN rmb_recharge.asset_id IS '关联的user_asset_detail_id';
COMMENT ON COLUMN rmb_recharge.tx_type IS '充值类型(0:ybt 1:kt)';
COMMENT ON COLUMN rmb_recharge.account IS '充值帐号';
COMMENT ON COLUMN rmb_recharge.subject IS '充值标题';
COMMENT ON COLUMN rmb_recharge.amount IS '充值金额';
COMMENT ON COLUMN rmb_recharge.txhash IS '交易id';
COMMENT ON COLUMN rmb_recharge.status IS '交易状态(-1:支付失败 0:支付中 1:成功)';
COMMENT ON COLUMN rmb_recharge.date IS '充值日期';
COMMENT ON COLUMN rmb_recharge.reason IS '交易状态原因';
COMMENT ON COLUMN rmb_recharge.over_time IS '过期时间';
COMMENT ON COLUMN rmb_recharge.extinfos IS '扩展信息';
COMMENT ON COLUMN rmb_recharge.create_time IS '注册时间';
COMMENT ON COLUMN rmb_recharge.update_time IS '更新时间';



-- 添加用户资产类型列表
CREATE TABLE "user_asset_type"
(
  id bigserial,
  user_id bigint, 
  "type" int default 0,
  total_amount double precision default 0,
  normal_amount double precision default 0,
  lock_amount double precision default 0,
  freeze_amount double precision default 0,    
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX user_asset_type_uidx ON "user_asset_type"(user_id, type);
ALTER TABLE user_asset_type REPLICA IDENTITY FULL;
COMMENT ON TABLE user_asset_type IS '用户每天ybt奖励明细表';
COMMENT ON COLUMN user_asset_type.user_id IS '用户id';
COMMENT ON COLUMN user_asset_type.type IS '货币类型';
COMMENT ON COLUMN user_asset_type.total_amount IS '总数量';
COMMENT ON COLUMN user_asset_type.normal_amount IS '可有数量';
COMMENT ON COLUMN user_asset_type.lock_amount IS '锁定数量';
COMMENT ON COLUMN user_asset_type.freeze_amount IS '冻结数量';
COMMENT ON COLUMN user_asset_type.create_time IS '注册时间';
COMMENT ON COLUMN user_asset_type.update_time IS '更新时间';




CREATE TABLE "currency_rate"
(
  id bigserial,	
	key varchar(32) default 0,	  
	"from" varchar(32) default 0,
  "to" varchar(32) default 0,
	ratio float default 0,
  source varchar(32) default '',
  digital boolean default false,
  auto boolean default true,
	create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX currency_rate_idx ON "currency_rate"(key);
ALTER TABLE currency_rate REPLICA IDENTITY FULL;
COMMENT ON TABLE currency_rate IS '货币汇率兑换表';
COMMENT ON COLUMN currency_rate.id IS '汇率id';
COMMENT ON COLUMN currency_rate.key IS '货币兑换key';
COMMENT ON COLUMN currency_rate.from IS '源币种';
COMMENT ON COLUMN currency_rate.to IS '目的币种种';
COMMENT ON COLUMN currency_rate.ratio IS '兑换比例';
COMMENT ON COLUMN currency_rate.source IS '汇率来源';
COMMENT ON COLUMN currency_rate.digital IS '是否数字货币';
COMMENT ON COLUMN currency_rate.auto IS '是否自动更新';
COMMENT ON COLUMN currency_rate.create_time IS '注册时间';
COMMENT ON COLUMN currency_rate.update_time IS '更新时间';

insert into currency_rate(key, "from", "to", create_time) values('hkdcny', 'hkd', 'cny', 1550644927);
insert into currency_rate(key, "from", "to", create_time) values('usdcny', 'usd', 'cny', 1550644927);





CREATE TABLE "voucher_info"
(
  id bigserial,		
	type smallint default 0,
	title varchar(64) default '',	
  context text default '',
  product_id bigint default '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX voucher_info_idx ON "voucher_info"(type);
ALTER TABLE voucher_info REPLICA IDENTITY FULL;
COMMENT ON TABLE voucher_info IS '代金券';
COMMENT ON COLUMN voucher_info.id IS '代金券id';
COMMENT ON COLUMN voucher_info.type IS '代金券币种';
COMMENT ON COLUMN voucher_info.title IS '代金券标题';
COMMENT ON COLUMN voucher_info.context IS '代金券内容';
COMMENT ON COLUMN voucher_info.product_id IS '代金券对应的商品id';
COMMENT ON COLUMN voucher_info.create_time IS '注册时间';
COMMENT ON COLUMN voucher_info.update_time IS '更新时间';


CREATE TABLE "voucher"
(
  id bigserial,
	user_id bigint default 0,
	type smallint default 0,  
	amount double PRECISION default 0,
	unlock_time int default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX voucher_idx ON "voucher"(user_id, type);
ALTER TABLE voucher REPLICA IDENTITY FULL;
COMMENT ON TABLE voucher IS '代金券';
COMMENT ON COLUMN voucher.id IS '代金券id';
COMMENT ON COLUMN voucher.user_id IS '所属用户';
COMMENT ON COLUMN voucher.type IS '代金券币种';
COMMENT ON COLUMN voucher.amount IS '代金券面值';
COMMENT ON COLUMN voucher.unlock_time IS '代金券过期时间';
COMMENT ON COLUMN voucher.create_time IS '注册时间';
COMMENT ON COLUMN voucher.update_time IS '更新时间';
select setval('voucher_id_seq', 5602556);   --设置id初始值

CREATE TABLE "voucher_record"
(
  id bigserial,	
  summary varchar(64) default 0,
	voucher_id bigint default 0,	
  to_uid bigint default 0,
	amount bigint default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX voucher_record_idx ON "voucher_record"(voucher_id);
ALTER TABLE voucher_record REPLICA IDENTITY FULL;
COMMENT ON TABLE voucher_record IS '代金券消费记录';
COMMENT ON COLUMN voucher_record.id IS '消费id';
COMMENT ON COLUMN voucher_record.summary IS '记录摘要';
COMMENT ON COLUMN voucher_record.to_uid IS '收款者用户id';
COMMENT ON COLUMN voucher_record.amount IS '付款金额';
COMMENT ON COLUMN voucher_record.create_time IS '注册时间';
COMMENT ON COLUMN voucher_record.update_time IS '更新时间';


create view ybasset_all as (select a.*,b.total_kt,b.total_kt_profit,b.total_issue_ybt,b.total_destroyed_ybt,b.total_mining,b.total_air_drop,b.total_air_unlock,b.total_activity,b.total_project,b.total_air_recover,b.total_perynbay from yunbay_asset_detail a join yunbay_asset b on a.date=b.date);