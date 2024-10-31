/**
CREATE USER ybapi WITH PASSWORD '123456';
CREATE DATABASE ybapi with owner=ybapi ENCODING='UTF8';
GRANT ALL PRIVILEGES ON DATABASE ybapi to ybapi;
\c ybapi ybapi;
**/


CREATE TABLE "user_asset"
(
  id bigserial,
  user_id bigint,  
  total_ybt_amount real default 0,
  total_kt_amount real default 0,
  lock_ybt_amount real default 0,
  lock_kt_amount real default 0, 
  create_time int NOT NULL,
  update_time int NOT NULL,
  primary key(id)
);

CREATE UNIQUE INDEX user_asset_user_idx ON "user_asset"(user_id);
ALTER TABLE user_asset REPLICA IDENTITY FULL;
COMMENT ON TABLE user_asset IS '用户资产表';
COMMENT ON COLUMN user_asset.user_id IS '用户id';
COMMENT ON COLUMN user_asset.total_ybt_amount IS '持有ybt总量';
COMMENT ON COLUMN user_asset.total_kt_amount IS '持有kt总量';
COMMENT ON COLUMN user_asset.lock_ybt_amount IS '锁定(冻结)ybt总量';
COMMENT ON COLUMN user_asset.lock_kt_amount IS '锁定(冻结)kt总量';
COMMENT ON COLUMN user_asset.create_time IS '注册时间';
COMMENT ON COLUMN user_asset.update_time IS '更新时间';

CREATE TABLE "user_asset_detail"
(
  id bigserial,
  user_id bigint,  
  currency_type smallint not null,
  transaction_type int not null,
  amount real not null,
  create_time int NOT NULL,
  update_time int NOT NULL,
  primary key(id)
);

CREATE INDEX user_asset_detail_userid_idx ON "user_asset_detail"(user_id, currency_type, transaction_type, create_time);
ALTER TABLE user_asset_detail REPLICA IDENTITY FULL;
COMMENT ON TABLE user_asset_detail IS '用户资产明细表';
COMMENT ON COLUMN user_asset_detail.user_id IS '用户id';
COMMENT ON COLUMN user_asset_detail.currency_type IS '货币类型(0:YBT 1:KT)';
COMMENT ON COLUMN user_asset_detail.transaction_type IS '交易类型(0:充币 1:交易挖矿 2:活动奖励 3:平台分红 4:提币)';
COMMENT ON COLUMN user_asset_detail.create_time IS '注册时间';
COMMENT ON COLUMN user_asset_detail.update_time IS '更新时间';


CREATE TABLE "yunbay_asset"
(
  id bigserial,
  amount real not null,
  profit real not null,
  issued real not null,
  canasset real not null,
  perynbay real not null,
  create_time int NOT NULL,
  update_time int NOT NULL,
  primary key(id)
);
CREATE INDEX yunbay_asset_createtime_idx ON yunbay_asset(create_time);
ALTER TABLE yunbay_asset REPLICA IDENTITY FULL;
COMMENT ON TABLE yunbay_asset IS '平台分红明细表';
COMMENT ON COLUMN yunbay_asset.amount IS '平台当日交易额';
COMMENT ON COLUMN yunbay_asset.issued IS '平台已发行云贝数';
COMMENT ON COLUMN yunbay_asset.canasset IS '可分红云贝数';
COMMENT ON COLUMN yunbay_asset.perynbay IS '每个云贝可得分红,如(1YBT=0.0001KT)';
COMMENT ON COLUMN yunbay_asset.create_time IS '注册时间';
COMMENT ON COLUMN yunbay_asset.update_time IS '更新时间';


CREATE TABLE "product"
(
  id bigserial,
  title varchar(256) not null DEFAULT '',
  info varchar(1024) DEFAULT '',
  images text[]  default NULL,
  canreturn boolean DEFAULT 'false',
  original real not null,
  sale real not null,
  rebat_percent real not null,
  inventory int DEFAULT 0,
  models bigint[] DEFAULT NULL,
  create_time int NOT NULL,
  update_time int NOT NULL,
  primary key(id)
);

CREATE INDEX product_createtime_idx ON product(create_time);
ALTER TABLE product REPLICA IDENTITY FULL;
COMMENT ON TABLE product IS '商品详情明细表';
COMMENT ON COLUMN product.title IS '商品标题';
COMMENT ON COLUMN product.info IS '商品描述';
COMMENT ON COLUMN product.images IS '商品图片';
COMMENT ON COLUMN product.canreturn IS '是否支持退换货';
COMMENT ON COLUMN product.original IS '原价';
COMMENT ON COLUMN product.sale IS '售价';
COMMENT ON COLUMN product.rebat_percent IS '贡献百分比';
COMMENT ON COLUMN product.inventory IS '库存量(-1:不限制库存)';
COMMENT ON COLUMN product.create_time IS '注册时间';
COMMENT ON COLUMN product.update_time IS '更新时间';


CREATE TABLE "product_model"
(
  id bigserial,
  product_id bigint,
  title varchar(256) not null DEFAULT '',
  info varchar(1024) DEFAULT '',
  captions varchar(128)[] DEFAULT NULL,
  original real not null,
  sale real not null,
  rebat_percent real not null,
  inventory int DEFAULT 0,
  create_time int NOT NULL,
  update_time int NOT NULL,
  primary key(id)
);
CREATE INDEX product_model_product_idx ON product_model(product_id);
ALTER TABLE product_model REPLICA IDENTITY FULL;
COMMENT ON TABLE product_model IS '商品规格明细表';
COMMENT ON COLUMN product_model.title IS '规格标题';
COMMENT ON COLUMN product_model.info IS '规格描述';
COMMENT ON COLUMN product_model.captions IS '标签';
COMMENT ON COLUMN product_model.original IS '原价';
COMMENT ON COLUMN product_model.sale IS '售价';
COMMENT ON COLUMN product_model.rebat IS '贡献百分比';
COMMENT ON COLUMN product_model.inventory IS '库存量(-1:不限制库存)';
COMMENT ON COLUMN product_model.create_time IS '注册时间';
COMMENT ON COLUMN product_model.update_time IS '更新时间';


CREATE TABLE "orders"
(
  id bigserial,
  user_id bigint not null,
  product_id bigint not null,
  product_model_id bigint default -1,
  quantity int DEFAULT 0,
  other_amount real DEFAULT 0,
  total_amount real not NULL,
  status int DEFAULT 0,
  rebat real DEFAULT 0,  
  create_time int NOT NULL,
  update_time int NOT NULL,
  primary key(id)
);
CREATE INDEX orders_userid_product_createtime_idx ON orders(user_id, product_id, create_time);
ALTER TABLE orders REPLICA IDENTITY FULL;
COMMENT ON TABLE orders IS '订单明细表';
COMMENT ON COLUMN orders.user_id IS '用户id';
COMMENT ON COLUMN orders.product_id IS '商品id';
COMMENT ON COLUMN orders.product_model_id IS '商品的规格id,-1为单个商品价格';
COMMENT ON COLUMN orders.quantity IS '购买数量';
COMMENT ON COLUMN orders.other_amount IS '其它扣费';
COMMENT ON COLUMN orders.total_amount IS '商品总价';
COMMENT ON COLUMN orders.status IS '订单状态(0:未付款 1:已付款 2:退款申请 3:已退款)';
COMMENT ON COLUMN orders.rebat IS '贡献值';
COMMENT ON COLUMN orders.create_time IS '注册时间';
COMMENT ON COLUMN orders.update_time IS '更新时间';

CREATE TABLE "feedback"
(
  id bigserial,
  user_id bigint not null,
  email varchar(256) not null,
  title varchar(1024) DEFAULT '',
  info text DEFAULT '',
  annex text DEFAULT '', 
  create_time int NOT NULL,
  update_time int NOT NULL,
  primary key(id)
);
CREATE INDEX feedback_userid_createtime_idx ON feedback(user_id, create_time);
ALTER TABLE feedback REPLICA IDENTITY FULL;
COMMENT ON TABLE feedback IS '用户反馈表';
COMMENT ON COLUMN feedback.user_id IS '用户id';
COMMENT ON COLUMN feedback.email IS '邮箱';
COMMENT ON COLUMN feedback.title IS '标题';
COMMENT ON COLUMN feedback.info IS '详情';
COMMENT ON COLUMN feedback.annex IS '附件';
COMMENT ON COLUMN feedback.create_time IS '注册时间';
COMMENT ON COLUMN feedback.update_time IS '更新时间';

CREATE TABLE "notice"
(
  id bigserial,
  user_id bigint DEFAULT 0,
  title varchar(1024) DEFAULT '',
  linkurl text DEFAULT '',
  create_time int NOT NULL,
  update_time int NOT NULL,
  primary key(id)
);

CREATE INDEX notice_createtime_idx ON notice(create_time);
ALTER TABLE notice REPLICA IDENTITY FULL;
COMMENT ON TABLE notice IS '平台公告表';
COMMENT ON TABLE notice.user_id IS '公告发布用户ID';
COMMENT ON COLUMN notice.title IS '标题';
COMMENT ON COLUMN notice.linkurl IS '跳转地址';
COMMENT ON COLUMN notice.create_time IS '注册时间';
COMMENT ON COLUMN notice.update_time IS '更新时间';


CREATE TABLE "banner"
(
  info text DEFAULT '',
  image text DEFAULT '', 
  linkurl text DEFAULT '',
  primary key(id)
)INHERITS(notice);

CREATE INDEX banner_createtime_idx ON banner(create_time);
ALTER TABLE banner REPLICA IDENTITY FULL;
COMMENT ON TABLE banner IS 'banner表';
COMMENT ON COLUMN banner.info IS '详情';
COMMENT ON COLUMN banner.image IS '图片url';
COMMENT ON COLUMN banner.linkurl IS '跳转地址';



CREATE TABLE "zixun"
(
  status smallint DEFAULT 0,
  primary key(id)
)INHERITS(notice);

CREATE INDEX zixun_status_idx ON zixun(status);
ALTER TABLE zixun REPLICA IDENTITY FULL;
COMMENT ON TABLE zixun IS '资讯动态表';
COMMENT ON COLUMN zixun.status IS '0:显示 1:下架';
