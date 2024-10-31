/**
CREATE USER ybapi WITH PASSWORD '123456';
CREATE DATABASE ybapi with owner=ybapi ENCODING='UTF8';
GRANT ALL PRIVILEGES ON DATABASE ybapi to ybapi;
ALTER DATABASE ybapi SET TIMEZONE='PRC';
\c ybapi ybapi;
**/


-- 创建array连接聚合
DROP AGGREGATE IF EXISTS anyarray_agg(anyarray);
CREATE AGGREGATE anyarray_agg(anyarray) (
  SFUNC = array_cat,
  STYPE = anyarray
);

CREATE FUNCTION unix_timestamp() RETURNS integer AS $$ 
--SELECT (date_part('epoch',now()))::integer;   
SELECT EXTRACT(epoch FROM NOW())::integer;
$$ LANGUAGE SQL IMMUTABLE;

CREATE FUNCTION from_unixtime(int) RETURNS timestamp AS $$ 
SELECT to_timestamp($1)::timestamp; 
$$ LANGUAGE SQL IMMUTABLE;



CREATE TABLE "invite"
(
  id bigserial,
  user_id bigint,  
  "type" SMALLINT DEFAULT 0,
  invite_userid bigint NOT NULL,
  invite_tel varchar(32) default '',
  recommend_userids bigint[] DEFAULT NULL,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX invite_userid_invite_userid_idx ON "invite"(user_id, invite_userid);
ALTER TABLE invite REPLICA IDENTITY FULL;
COMMENT ON TABLE invite IS '邀请详情表';
COMMENT ON COLUMN invite.user_id IS '用户id';
COMMENT ON COLUMN invite.type IS '邀请还是被邀请';
COMMENT ON COLUMN invite.invite_userid IS '邀请人';
COMMENT ON COLUMN invite.recommend_userids IS '推荐人列表';
COMMENT ON COLUMN invite.user_id IS '用户id';
COMMENT ON COLUMN invite.invite_tel IS '邀请人号码';
COMMENT ON COLUMN invite.create_time IS '注册时间';
COMMENT ON COLUMN invite.update_time IS '更新时间';


DROP TABLE IF EXISTS product_category;
CREATE TABLE "product_category"
(
  id bigserial,
  title varchar(256) NOT NULL DEFAULT '',
  info varchar(500) DEFAULT '',
  admin_user_id bigint DEFAULT 0,
  parent_id bigint DEFAULT 0,
  picture varchar(255) DEFAULT '',
  sort int default 0,
  is_show smallint default 1,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX uidx_product_category ON "product_category" USING btree ("title", "parent_id");
COMMENT ON TABLE product_category IS '商品分类表';
COMMENT ON COLUMN product_category.title IS '商品分类标题';
COMMENT ON COLUMN product_category.info IS '分类详情';
COMMENT ON COLUMN product_category.admin_user_id IS '操作用户ID';
COMMENT ON COLUMN product_category.parent_id IS '父级分类id';
COMMENT ON COLUMN product_category.picture IS '分类图片';
COMMENT ON COLUMN product_category.sort IS '排序';
COMMENT ON COLUMN product_category.is_show IS '是否展示；1：是，0：否（不展示）';
COMMENT ON COLUMN product_category.create_time IS '创建时间';
COMMENT ON COLUMN product_category.update_time IS '更新时间';


DROP TABLE IF EXISTS product;
CREATE TABLE "product"
(
  id bigserial,
  category_id bigint default 0,
  user_id bigint default 0,
  title varchar(256) not null DEFAULT '',
  info varchar(1024) DEFAULT '',
  images text[]  default NULL,
  descimgs jsonb DEFAULT NULL,
  "type" smallint default 0,
  canreturn boolean DEFAULT false,
  stock int not null default 0,  
  sold int not null DEFAULT 0,
  cost_price numeric not null default 0,
  price numeric not null default 0,
  rebat numeric not null default 0,
  contact jsonb DEFAULT NULL,
  def_sku_id bigint default 0,
  --pay_type varchar(12)[] default '{}',
  is_hid smallint not null default 0,
  hid_cause varchar default '',
  check_status smallint default 1,
  publish_area smallint default 0,
  reason varchar default '',
  country smallint default 0,
  extinfo jsonb not null default '{}',
  status smallint DEFAULT 1,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX idx_product ON product(publish_area, check_status, status, update_time);
CREATE INDEX idx_product_type ON product(category_id, user_id, "type");

COMMENT ON TABLE product IS '商品详情明细表';
COMMENT ON COLUMN product.category_id IS '商品所属分类id';
COMMENT ON COLUMN product.user_id IS '商品所属用户id';
COMMENT ON COLUMN product.title IS '商品标题';
COMMENT ON COLUMN product.info IS '特色介绍';
COMMENT ON COLUMN product.images IS '商品图片';
COMMENT ON COLUMN product.descimgs IS '商品描述图';
COMMENT ON COLUMN product.type IS '商品类型(0:实物 1:虚拟(话费充值) 2:虚拟(点卡))';
COMMENT ON COLUMN product.canreturn IS '是否支持退换货，默认false：不支持，true：支持';
COMMENT ON COLUMN product.stock IS '总库存';
COMMENT ON COLUMN product.sold IS '已售总量(所有规格之和)';
COMMENT ON COLUMN product.price IS '商品价格';
COMMENT ON COLUMN product.rebat IS '贡献百分比';
COMMENT ON COLUMN product.contact IS '售后联系方式';
COMMENT ON COLUMN product.def_sku_id_sku_id IS '默认选中的商品规格id';
COMMENT ON COLUMN product.status IS '上下架(1:上架 ;0:下架)';
COMMENT ON COLUMN product.is_hid IS '是否屏蔽(0:未屏蔽 1:已屏蔽)';
COMMENT ON COLUMN product.check_status IS '审核状态(-1:不通过, 0:待审核 1:审核通过)';
COMMENT ON COLUMN product.create_time IS '创建时间';
COMMENT ON COLUMN product.update_time IS '更新时间';



DROP TABLE IF EXISTS product_sku;
CREATE TABLE "product_sku"
(
  id bigserial,
  product_id bigint,
  sku jsonb not null DEFAULT '{}',
  combines jsonb not null default '[]',
  stock int not null default 0,
  sold int not null DEFAULT 0,
  cost_price numeric not null default 0,
  price numeric not null default 0,
  img text not null default '',
  extinfo jsonb default '{}',    
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX idx_product_sku ON product_sku(product_id);
ALTER TABLE product_sku REPLICA IDENTITY FULL;
COMMENT ON TABLE product_sku IS '商品规格明细表';
COMMENT ON COLUMN product_sku.img IS '规格图片';
COMMENT ON COLUMN product_sku.cost_price IS '成本价';
COMMENT ON COLUMN product_sku.price IS '售价';
COMMENT ON COLUMN product_sku.stock IS '库存量(0:不限制库存)';
COMMENT ON COLUMN product_sku.sold IS '已售量';
COMMENT ON COLUMN product_sku.extinfo IS '扩展信息';
COMMENT ON COLUMN product_sku.create_time IS '注册时间';
COMMENT ON COLUMN product_sku.update_time IS '更新时间';



DROP TABLE IF EXISTS product_attr_key;
CREATE TABLE "product_attr_key"
(
  id bigserial,
  category_id bigint,
  name varchar not null,  
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX uidx_product_attr_key ON product_attr_key(category_id, name);
ALTER TABLE product_attr_key REPLICA IDENTITY FULL;
COMMENT ON TABLE product_attr_key IS '商品属性key表';
COMMENT ON COLUMN product_attr_key.category_id IS '所属商品分类id';
COMMENT ON COLUMN product_attr_key.name IS '属性key';
COMMENT ON COLUMN product_attr_key.create_time IS '注册时间';
COMMENT ON COLUMN product_attr_key.update_time IS '更新时间';


DROP TABLE IF EXISTS product_attr_value;
CREATE TABLE "product_attr_value"
(
  id bigserial,
  product_attr_key_id bigint,
  value varchar not null,  
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX uidx_product_attr_value ON product_attr_value(product_attr_key_id, value);
ALTER TABLE product_attr_value REPLICA IDENTITY FULL;
COMMENT ON TABLE product_attr_value IS '商品属性值表';
COMMENT ON COLUMN product_attr_value.value IS '属性value';
COMMENT ON COLUMN product_attr_value.create_time IS '注册时间';
COMMENT ON COLUMN product_attr_value.update_time IS '更新时间';


DROP TABLE IF EXISTS product_price;
CREATE TABLE "product_price"
(
  id bigserial,
  p_id bigint not null default 0,
  p_sku_id bigint not null default 0,
  stock int not null default 0,
  sold int not null default 0,
  cost_price numeric not null default 0,
  price numeric not null default 0,  
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX uidx_p_price ON product_price(p_id, p_sku_id);
ALTER TABLE product_price REPLICA IDENTITY FULL;
COMMENT ON TABLE product_price IS '商品价格表';
COMMENT ON COLUMN product_price.p_id IS '商品id';
COMMENT ON COLUMN product_price.p_sku_id IS '商品规格id';
COMMENT ON COLUMN product_price.stock IS '库存';
COMMENT ON COLUMN product_price.sold IS '已售';
COMMENT ON COLUMN product_price.cost_price IS '原价';
COMMENT ON COLUMN product_price.price IS '售价';
COMMENT ON COLUMN product_price.create_time IS '注册时间';
COMMENT ON COLUMN product_price.update_time IS '更新时间';




DROP TABLE IF EXISTS product_recommend;
CREATE TABLE "product_recommend"
(
  id bigserial,
  "type" int default 0,
  "name" varchar(64) default '',
  img text DEFAULT '',
  descimg text DEFAULT '',
  product_ids bigint[] not null,  
  country smallint not null default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX product_recommend_type_idx ON product_recommend(type);
ALTER TABLE product_recommend REPLICA IDENTITY FULL;
COMMENT ON TABLE product_recommend IS '商品推荐表';
COMMENT ON COLUMN product_recommend.type IS '推荐类型(0:精选 1:最新)';
COMMENT ON COLUMN product_recommend.name IS '类型名称';
COMMENT ON COLUMN product_recommend.img IS '封面';
COMMENT ON COLUMN product_recommend.descimg IS '更多封面';
COMMENT ON COLUMN product_recommend.product_ids IS '推荐的商品类型id列表';
COMMENT ON COLUMN product_recommend.country IS '地区(0:国内 1:国外)';
COMMENT ON COLUMN product_recommend.create_time IS '注册时间';
COMMENT ON COLUMN product_recommend.update_time IS '更新时间';

CREATE TABLE "logistics"
(
  id bigserial,
  order_id bigint not null,
  user_id bigint not null,
  company varchar(64) default '',
  number varchar(64) default '',
  infos text[] DEFAULT NULL,
  status smallint DEFAULT NULL,  
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX logistics_userid_number_createtime_idx ON logistics(user_id, number, create_time);
ALTER TABLE logistics REPLICA IDENTITY FULL;
COMMENT ON TABLE logistics IS '物流表';
COMMENT ON COLUMN logistics.order_id IS '订单id';
COMMENT ON COLUMN logistics.user_id IS '用户id';
COMMENT ON COLUMN logistics.company IS '物流公司';
COMMENT ON COLUMN logistics.number IS '物流单号';
COMMENT ON COLUMN logistics.infos IS '物流信息';
COMMENT ON COLUMN logistics.status IS '物流状态(0:未完成 1:已完成)';
COMMENT ON COLUMN logistics.create_time IS '注册时间';
COMMENT ON COLUMN logistics.update_time IS '更新时间';


CREATE TABLE "cart"
(
  id bigserial,
  user_id bigint not null,
  seller_userid bigint default 0,
  product_id bigint not null,
  product_sku_id bigint default -1,  
  quantity int DEFAULT 0,
  other_amount numeric DEFAULT 0,
  product jsonb DEFAULT NULL,
  country smallint default 0,
  publish_area smallint default 1,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);

CREATE UNIQUE INDEX cart_user_idx ON cart(user_id, product_id, product_sku_id);
ALTER TABLE cart REPLICA IDENTITY FULL;
COMMENT ON TABLE cart IS '购物车表';
COMMENT ON COLUMN cart.user_id IS '用户id';
COMMENT ON COLUMN cart.product_id IS '商品id';
COMMENT ON COLUMN cart.product_sku_id IS '商品的规格id';
COMMENT ON COLUMN cart.quantity IS '购买数量';
COMMENT ON COLUMN cart.other_amount IS '其它扣费(运费)';
COMMENT ON COLUMN cart.product IS '商品结构信息';
COMMENT ON COLUMN cart.create_time IS '注册时间';
COMMENT ON COLUMN cart.update_time IS '更新时间';
comment on column cart.publish_area is '销售专区；默认1：KT， 0：YBT';
COMMENT ON COLUMN cart.country IS '国家(0:国际版 1:国内版)';




CREATE TABLE "orders"
(
  id bigserial,
  user_id bigint not null,
  seller_userid bigint default 0,
  product_id bigint not null,
  product_sku_id bigint default 0,  
  address_info jsonb default null,
  logistics_id bigint DEFAULT 0,
  quantity int DEFAULT 0,
  currency_type smallint default 1,  
  currency_percent NUMERIC DEFAULT 0,
  other_amount NUMERIC DEFAULT 0,
  rebat_amount NUMERIC DEFAULT 0,
  total_amount NUMERIC DEFAULT 0,  
  status int DEFAULT 0,  
  sale_status smallint DEFAULT 0,
  product jsonb DEFAULT NULL,
  extinfos jsonb DEFAULT '{}',
  auto_cancel_time int DEFAULT 0,
  auto_finish_time int DEFAULT 0,  
  date varchar(20) default '',
  publish_area smallint default 1,
  maninfos jsonb not null default '{}',
  auto_deliver boolean DEFAULT false,
  shield smallint default 0,
  country smallint default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);

CREATE INDEX orders_user_idx ON orders(user_id, seller_userid, status, create_time);
CREATE INDEX orders_product_idx ON orders(product_id, product_sku_id, date, shield);
CREATE INDEX orders_man_idx ON orders(publish_area, maninfos);
ALTER TABLE orders REPLICA IDENTITY FULL;
COMMENT ON TABLE orders IS '订单明细表';
COMMENT ON COLUMN orders.user_id IS '用户id';
COMMENT ON COLUMN orders.product_id IS '商品id';
COMMENT ON COLUMN orders.product_sku_id IS '商品的规格id';
COMMENT ON COLUMN orders.quantity IS '购买数量';
COMMENT ON COLUMN orders.currency_type IS '货币类型(0:ybt 1:kt)';
COMMENT ON COLUMN orders.currency_percent IS 'rmb兑换货币比例';
COMMENT ON COLUMN orders.other_amount IS '其它扣费(运费)';
COMMENT ON COLUMN orders.rebat_amount IS '订单贡献值';
COMMENT ON COLUMN orders.total_amount IS '订单总价';
COMMENT ON COLUMN orders.status IS '订单状态(0:购物车 1:未付款 2:已付款 3:已发货 4:已完成 5:已取消)';
COMMENT ON COLUMN orders.sale_status IS '售后状态(0:未售后 1:售后中 2:售后完成)';
COMMENT ON COLUMN orders.product IS '商品结构快照信息(待付款后)';
COMMENT ON COLUMN orders.extinfos IS '扩展信息';
COMMENT ON COLUMN orders.address_info IS '地址信息';
COMMENT ON COLUMN orders.logistics_id IS '物流id';
COMMENT ON COLUMN orders.auto_cancel_time IS '订单自动取消时间(未付款时有效)';
COMMENT ON COLUMN orders.auto_finish_time IS '订单自动确认收货时间(当前状态为已发货时有效)';
COMMENT ON COLUMN orders.date IS '订单生成日期(待支付日期)';
COMMENT ON COLUMN orders.shield IS '是否屏蔽( 0:未屏蔽 1:已屏蔽)';
COMMENT ON COLUMN orders.create_time IS '注册时间';
COMMENT ON COLUMN orders.update_time IS '更新时间';
comment on column orders.publish_area is '销售专区；默认1：KT， 0：YBT';
comment on column orders.maninfos is '后台扩展字段';
COMMENT ON COLUMN orders.auto_deliver IS '订单是否自动发货';
COMMENT ON COLUMN orders.country IS '国家(0:国际版 1:国内版)';
select setval('orders_id_seq', 1602556);   --设置订单id初始值







CREATE TABLE "orders_status"
(
  id bigserial,
  order_id bigint not null,
  "status" smallint default 0,
  date varchar(20) default '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX orders_status_idx ON orders_status(order_id, status);
ALTER TABLE orders_status REPLICA IDENTITY FULL;
COMMENT ON TABLE orders_status IS '订单状态信息表';
COMMENT ON COLUMN orders_status.order_id IS '订单id';
COMMENT ON COLUMN orders_status.status IS '订单状态';
COMMENT ON COLUMN orders_status.date IS '订单状态日期';
COMMENT ON COLUMN orders_status.create_time IS '注册时间';
COMMENT ON COLUMN orders_status.update_time IS '更新时间';


DROP TABLE IF EXISTS business;
CREATE TABLE "business"
(
  id bigserial,
  user_id bigint not null,
  tel varchar(32) default '',
  type smallint default 0,
  company varchar(512) default '',
  license text[] default null,
  name varchar(64) default '',
  location varchar(32) default '',
  certype smallint default 0,
  certid varchar(64) default '',
  certimgs  text[] default null,
  product_types bigint[] default null,
  hasbusiness boolean default false,
  website text default '',
  contact jsonb default null,
  status smallint default 0,
  not_pass_cause varchar(500) default '',
  total_tradeflow NUMERIC default 0,
  total_rebat NUMERIC default 0,
  is_ybt smallint default 0,
  country smallint default 0,
  total_ybtflow NUMERIC default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX business_uniqueidx ON business(user_id);
CREATE INDEX business_idx ON business(type, status);
ALTER TABLE business REPLICA IDENTITY FULL;
COMMENT ON TABLE business IS '商家信息表';
COMMENT ON COLUMN business.user_id IS '用户id';
COMMENT ON COLUMN business.tel IS '注册手机号';
COMMENT ON COLUMN business.type IS '类型(0:个人 1:企业)';
COMMENT ON COLUMN business.company IS '企业名称';
COMMENT ON COLUMN business.license IS '营业执照';
COMMENT ON COLUMN business.name IS '负责人';
COMMENT ON COLUMN business.location IS '所在地';
COMMENT ON COLUMN business.certype IS '证件类型（0：身份证，1：护照）';
COMMENT ON COLUMN business.certid IS '证件id';
COMMENT ON COLUMN business.certimgs IS '证件图片';
COMMENT ON COLUMN business.product_types IS '经营范围（一级分类id）';
COMMENT ON COLUMN business.hasbusiness IS '是否有在线电商平台（false：无，true：有）';
COMMENT ON COLUMN business.website IS '电商平台网址';
COMMENT ON COLUMN business.contact IS '联系方式';
COMMENT ON COLUMN business.status IS '审核状态(0:未审核 1:审核通过 2:审核不通过)';
COMMENT ON COLUMN business.not_pass_cause IS '不通过原因';
COMMENT ON COLUMN business.total_tradeflow IS '总交易流水';
COMMENT ON COLUMN business.total_rebat IS '总贡献值';
COMMENT ON COLUMN business.create_time IS '注册时间';
COMMENT ON COLUMN business.update_time IS '更新时间';
comment on column business.total_ybtflow is '商家ybt交易额';
comment on column business.country is '国家（0:国际版, 1:国内版）';

CREATE TABLE "user_address"
(
  id bigserial,
  user_id bigint not null,
  receiver varchar(64) not null,
  tel varchar(32) not null,
  address varchar(512)[] not null,
  "default" boolean default true,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX user_address_userid_idx ON user_address(user_id);
ALTER TABLE user_address REPLICA IDENTITY FULL;
COMMENT ON TABLE user_address IS '用户地址表';
COMMENT ON COLUMN user_address.user_id IS '用户id';
COMMENT ON COLUMN user_address.receiver IS '收货人';
COMMENT ON COLUMN user_address.tel IS '收货人电话';
COMMENT ON COLUMN user_address.address IS '收货地址';
COMMENT ON COLUMN user_address.default IS '是否默认地址';
COMMENT ON COLUMN user_address.create_time IS '注册时间';
COMMENT ON COLUMN user_address.update_time IS '更新时间';


CREATE TABLE "feedback"
(
  id bigserial,
  user_id bigint not null,
  email varchar(256) not null,
  title varchar(1024) DEFAULT '',
  info text DEFAULT '',
  affix text[] DEFAULT NULL,
  type smallint default 1,
  status smallint default 0,
  reply_admin_user_id bigint default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX feedback_userid_createtime_idx ON feedback(user_id, create_time);
ALTER TABLE feedback REPLICA IDENTITY FULL;
COMMENT ON TABLE feedback IS '用户反馈表';
COMMENT ON COLUMN feedback.user_id IS '用户id';
COMMENT ON COLUMN feedback.email IS '邮箱';
COMMENT ON COLUMN feedback.title IS '标题';
COMMENT ON COLUMN feedback.info IS '详情';
COMMENT ON COLUMN feedback.affix IS '附件';
COMMENT ON COLUMN feedback.type IS '类型；默认1：故障反馈，2：买家申述，3：卖家申述';
COMMENT ON COLUMN feedback.status IS '状态；默认0：未回复，1：已回复';
COMMENT ON COLUMN feedback.reply_admin_user_id IS '回复人员id';
COMMENT ON COLUMN feedback.create_time IS '注册时间';
COMMENT ON COLUMN feedback.update_time IS '更新时间';

CREATE TABLE "notice"
(
  id bigserial,
  type smallint DEFAULT 0,
  user_id bigint DEFAULT 0,
  title varchar(1024) DEFAULT '',
  linkurl text DEFAULT '',
  context text DEFAULT '',
  status smallint DEFAULT 0,  
  create_time int default 0,
  country smallint default 0,
  update_time int default 0,
  primary key(id)
);

CREATE INDEX notice_createtime_idx ON notice(type, create_time);
ALTER TABLE notice REPLICA IDENTITY FULL;
COMMENT ON TABLE notice IS '平台公告表';
COMMENT ON COLUMN notice.type IS '类型(0:平台公告 1:常见问题 2:资讯动态)';
COMMENT ON COLUMN notice.user_id IS '发布用户ID';
COMMENT ON COLUMN notice.title IS '标题';
COMMENT ON COLUMN notice.linkurl IS '跳转地址';
COMMENT ON COLUMN notice.context IS '内容';
COMMENT ON COLUMN notice.status IS '状态(-1:下线 0:发布状态 1:推荐状态)';
COMMENT ON COLUMN notice.create_time IS '创建时间';
COMMENT ON COLUMN notice.update_time IS '更新时间';
comment on column notice.country is '国家（0:国际版, 1:国内版）';


DROP TABLE IF EXISTS banner;
CREATE TABLE "banner"
(
  id bigserial,
  position smallint DEFAULT 1,
  content jsonb DEFAULT null,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX banner_position_index ON banner(position);
COMMENT ON TABLE banner IS 'banner表';
COMMENT ON COLUMN banner.position IS '位置，默认1：移动端app首页';
COMMENT ON COLUMN banner.content IS '内容';
COMMENT ON COLUMN banner.create_time IS '创建时间';
COMMENT ON COLUMN banner.update_time IS '更新时间';


drop table if exists upgrade;
CREATE TABLE upgrade (
    id serial,
    type integer NOT NULL DEFAULT 0,
    platform varchar(64) NOT NULL DEFAULT '',
    version varchar(64) NOT NULL DEFAULT '',
    verint bigint NOT NULL DEFAULT 0,
    url varchar(512) NOT NULL DEFAULT '',
    md5 varchar(64) NOT NULL DEFAULT '',
    upversions varchar(64)[] NULL,
    channels varchar(64)[] NULL,
    status integer NOT NULL DEFAULT 0,
    title varchar(128) NOT NULL DEFAULT '',
    "desc" varchar(512) NOT NULL DEFAULT '',
    maner varchar(64) NOT NULL DEFAULT 0,
    mandatory jsonb default null,
    country smallint default 0,
    create_time integer NULL DEFAULT 0,
    update_time integer NULL DEFAULT 0,
    primary key (id)
);
CREATE INDEX idx_upgrade_platform on upgrade(platform);
CREATE INDEX idx_upgrade_verint on upgrade(verint);
COMMENT ON TABLE upgrade IS '升级配置表';
COMMENT ON COLUMN upgrade.type IS '升级方式(0:非弹窗通知升级 1:弹窗非强升 2:强窗强升)';
COMMENT ON COLUMN upgrade.version IS '升级包版本号';
COMMENT ON COLUMN upgrade.verint IS '需要升级的版本号整形, 值为指定的版本或者当前升级包版本';
COMMENT ON COLUMN upgrade.platform IS '升级包的平台(androis,ios)';
COMMENT ON COLUMN upgrade.url IS '升级包url';
COMMENT ON COLUMN upgrade.md5 IS '升级包md5';
COMMENT ON COLUMN upgrade.upversions IS '指定升级的版本号,空为所有老版本';
COMMENT ON COLUMN upgrade.channels IS '指定升级的渠道,空为所有渠道';
COMMENT ON COLUMN upgrade.status IS '状态 0:下架 1:上架';
COMMENT ON COLUMN upgrade.title IS '更新标题';
COMMENT ON COLUMN upgrade.desc IS '更新说明';
COMMENT ON COLUMN upgrade.maner IS '创建人';
COMMENT ON COLUMN upgrade.create_time IS '创建时间';
COMMENT ON COLUMN upgrade.update_time IS '更新时间';
comment on column upgrade.mandatory is '强制升级版本结构';
comment on column upgrade.country is '国家(0:国际 1:中国)';


-- 创建订单状态触发器 记录订单不同状态时间点
-- 创建订单状态触发器函数
CREATE OR REPLACE FUNCTION fn_orders_status_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  obj record;
  _date varchar(20);
  now int;
  _orderid bigint;
  _status int;
BEGIN
    -- 记录订单状态信息
    -- IF (TG_OP = 'DELETE') THEN    --删除记录只有OLD，新增和更新是NEW
    --   _status = OLD.status;
    --   _orderid = OLD.id;

    --   delete from orders_status where order_id=_orderid and status=_status;
    -- ELSE    
      if NEW.status > 0 then
        _status = NEW.status;
        _orderid = NEW.id;
        _date = current_date;
        now = unix_timestamp();

        insert into orders_status(order_id, status, date, create_time, update_time) values(_orderid, _status, _date, now, now) on conflict(order_id, status) do nothing;
      end if;
    -- END IF;

    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建用订单状态触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_orders_status on orders;
CREATE TRIGGER tg_to_update_orders_status
after INSERT OR UPDATE of status    -- 前触发
--BEFORE UPDATE 
ON orders  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_orders_status_value();


-- 创建商品规格销量触发器
-- 创建商品规格销量触发器函数 
CREATE OR REPLACE FUNCTION fn_product_model_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  _product_id bigint;
  obj record;
  product_model_row record;
BEGIN
    -- 销量
    IF (TG_OP = 'DELETE') THEN    --删除记录只有OLD，新增和更新是NEW      
      _product_id = OLD.product_id;      
    ELSE    
      _product_id = NEW.product_id;            
    END IF;

    select sum(sold_quantity) as total_sold_quantity, sum(quantity) as total_quantity into obj from product_model where product_id=_product_id;
    select sum(quantity) - sum(sold_quantity) as total_quantity into product_model_row from product_model where product_id=_product_id and quantity<>0;
    if obj is not null then
      if product_model_row.total_quantity is null then
        product_model_row.total_quantity =0;
      end if;
      update product set total_sold_quantity=obj.total_sold_quantity, total_quantity=product_model_row.total_quantity where id=_product_id;
    end if;
    
    RETURN null;        
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建商品规格销量触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_product_model on product_model;
CREATE TRIGGER tg_to_update_product_model
after INSERT OR UPDATE of quantity,sold_quantity    -- 后触发
ON product_model  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_product_model_value();


drop table if exists admin_user;
create table admin_user (
  id bigserial,
  username varchar(50) not null default '',
  nickname varchar(50) default '',
  email varchar(100) default '',
  password char(64) default '',
  role_id bigint default 0,
  status smallint default 1,
  create_time int default 0,
  update_time int default 0,
  remark varchar(250) default '',
  primary key(id)
);
create unique index username_index on admin_user using btree (username);
comment on table admin_user is '后台用户表';
comment on column admin_user.username is '用户名';
comment on column admin_user.nickname is '用户昵称';
comment on column admin_user.email is '电子邮箱';
comment on column admin_user.password is '密码';
comment on column admin_user.role_id is '角色id';
comment on column admin_user.status is '状态：1：正常，0：已禁用';
comment on column admin_user.create_time is '创建时间';
comment on column admin_user.update_time is '更新时间';


-- user_action 用户权限表
drop table if exists admin_user_action;
create table admin_user_action (
  id bigserial,
  admin_user_id bigint default 0,
  own_url_path varchar(100)[] default '{}',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
create unique index user_id_index on admin_user_action using btree("admin_user_id");
comment on table admin_user_action is '用户权限表';
comment on column admin_user_action.admin_user_id is '用户id';
comment on column admin_user_action.own_url_path is '拥有的权限';
comment on column admin_user_action.create_time is '创建时间';
comment on column admin_user_action.update_time is '更新时间';

-- action权限表
drop table if exists admin_action;
create table admin_action(
  id bigserial,
  controller_id bigint default 0,
  action varchar(80) default '',
  url_path varchar(100) default '',
  name varchar(200) default '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
create unique index url_path_index on admin_action using btree("url_path");
comment on table admin_action is '权限表';
comment on column admin_action.controller_id is '大功能id';
comment on column admin_action.action is '行为名';
comment on column admin_action.url_path is '权限（访问路径）';
comment on column admin_action.name is '权限名';
comment on column admin_action.create_time is '创建时间';
comment on column admin_action.update_time is '更新时间';

-- controller 大功能点
drop table if exists admin_controller;
create table admin_controller(
  id bigserial,
  module varchar(50) default 'backend',
  controller varchar(80) default '',
  name varchar(100) default '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
comment on table admin_controller is '大功能表';
comment on column admin_controller.module is '模块名';
comment on column admin_controller.controller is '控制器名';
comment on column admin_controller.name is '功能名';
comment on column admin_controller.create_time is '创建时间';
comment on column admin_controller.update_time is '更新时间';



-- vote_bourse 投票上交易所表
drop table if exists vote_bourse;
CREATE TABLE vote_bourse (
    id serial,
    type integer NOT NULL DEFAULT 1,
    user_id bigint default 0,
    create_time integer NULL DEFAULT 0,
    update_time integer NULL DEFAULT 0,
    primary key (id)
);
CREATE INDEX idx_user_id on vote_bourse(user_id);
COMMENT ON TABLE vote_bourse IS '投票上交易所表';
COMMENT ON COLUMN vote_bourse.type IS '时间选项(1:挖矿3天， 2:挖矿5天， 3:挖矿7天，4:挖矿15天， 5:更长时间)';
COMMENT ON COLUMN vote_bourse.user_id IS '用户id';
COMMENT ON COLUMN vote_bourse.create_time IS '创建时间';
COMMENT ON COLUMN vote_bourse.update_time IS '更新时间';



-- 新增搜索增量标识计数表
DROP TABLE IF EXISTS sphinx_counter;
CREATE TABLE "sphinx_counter"
(
  id bigserial,
  table_name varchar(50) default '',
  min_id bigint default 0,
  max_id bigint default 0,
  max_time bigint default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX table_name_index ON "sphinx_counter" using btree ("table_name");
COMMENT ON TABLE sphinx_counter IS '搜索增量标识计数表';
COMMENT ON COLUMN sphinx_counter.table_name is '表名';
COMMENT ON COLUMN sphinx_counter.min_id is '最小id';
COMMENT ON COLUMN sphinx_counter.max_id is '最大id';
COMMENT ON COLUMN sphinx_counter.max_time is '最大修改时间';
COMMENT ON COLUMN sphinx_counter.create_time is '创建时间';
COMMENT ON COLUMN sphinx_counter.update_time is '更新时间';

insert into sphinx_counter (table_name, create_time) values ('product', 1543224177); 




-- 站点设置表
drop table if exists setting;
create table setting (
    id serial,
    setting_key varchar(255) default '',
    setting_value text default '',
    create_time integer default 0,
    update_time integer default 0,
    primary key (id)
);
create unique index index_setting_key on setting(setting_key);
comment on table setting is '站点设置表';
comment on column setting.setting_key is '设置key';
comment on column setting.setting_value is '设置值';
comment on column setting.create_time is '创建时间';
comment on column setting.update_time is '更新时间';


INSERT INTO "public"."admin_action" VALUES ('80', '6', 'GetDrawLimit', '/backend/setting/get-draw-limit', '提币免审限额设置 列表', '1541385296', '1541385296');
INSERT INTO "public"."admin_action" VALUES ('81', '6', 'SaveDrawLimit', '/backend/setting/save-draw-limit', '提币免审限额设置 修改', '1541385511', '1541385511');
INSERT INTO "public"."admin_action" VALUES ('83', '6', 'CommonAccount', '/backend/wallet/common-account', '公共账号钱包 列表', '1541385511', '1541385511');
INSERT INTO "public"."admin_action" VALUES ('84', '6', 'RemarkAccount', '/backend/wallet/remark-account', '公共账号钱包 编辑备注', '1541385511', '1541385511');


-- 公共账号钱包表
drop table if exists common_account_address;
create table common_account_address (
    id serial,
    name varchar(200) default '',
    uid varchar(100) default '',
    address varchar(80) default '',
    type smallint default 0,
    remark varchar(255) default '',
    create_time integer default 0,
    update_time integer default 0,
    primary key(id)
);

create unique index index_address on common_account_address(address);
comment on table common_account_address is '公共账号钱包表';
comment on column common_account_address.name is '名称';
comment on column common_account_address.uid is 'yunbay账号ID';
comment on column common_account_address.address is '地址';
comment on column common_account_address.type is '平台；0：内盘地址，1：外部地址';
comment on column common_account_address.remark is '备注';
comment on column common_account_address.create_time is '创建时间';
comment on column common_account_address.update_time is '更新时间';

insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (1, '提币节点钱包', '', '0xd13c24341178abd8144eabe4431752e0199b6956', 1, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (2, '系统', '0', '0xef45c74bd04738819501b8a2ec78d120467d075d', 0, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (3, '团队激励10%', '1', '0x7e72db7e2dedbc4e874f193332cab4c87734b0b7', 0, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (4, '研发运营13%', '2', '0x09e790db5b17fa9e2c8ff0774d2f56bc1f9cef89', 0, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (5, '回购5%', '3', '0xc57c20a9022bd6306516c6275b390f5b05013ad4', 0, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (6, '商家质量保证金', '8', '0x5cf3b5db9129e12ef8e53fb5588e8a8b9369f807', 0, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (7, '外部钱包分红', '10', '0x27a58d2575c4e96b1a841b4927c06edcda244a43', 0, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (8, '热币转入云贝分配', '21', '0x817d122e2f9983251280a04a5756e980eff7ac60', 0, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (9, '热币转入云贝回收', '22', '0xbd0e511feef3920a31ed3e83a98f5a16cd143db1', 0, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (10, '云贝转入热币分配', '', '0x958a506e14073efed7a5551baa8fdc06b2070e72', 1, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (11, '提币到yunEX分红', '23', '0x4a48f4638385288f0b88a69223af3da8a2eff39c', 0, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (12, '云贝转入yunEX回收', '24', '0xb826d3a4bc61f4a8f5a451850b5cf81a774c77f9', 0, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (13, 'yunEX转入云贝分配（云网所有）', '55254', '0x827371e3887c5fb71f9aad410df4ba9d78988f72', 0, 1541385511, 1541385511);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (14, '云贝转入yunEX分配', '', '0xa3de156e70a37793a838ed0b9f1fd3170276219f', 1, 1541385511, 1541385511);





drop table "of_order";
CREATE TABLE "of_order"
(
  id bigserial,
  order_id bigint not null,
  of_id varchar(32) not null default '',
  cardid varchar(32) default '',
  cardname varchar(32) default '',
  cardnum int default 0, 
  ordercash numeric default 0,
  game_userid varchar(32) default '',
  game_state int default 0,  
  reason varchar(64) default '',
  retcode smallint default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX idx_of_order ON of_order(order_id, of_id);
ALTER TABLE of_order REPLICA IDENTITY FULL;
COMMENT ON TABLE of_order IS '欧飞订单表';
COMMENT ON COLUMN of_order.order_id IS '订单id';
COMMENT ON COLUMN of_order.of_id IS '欧飞订单id';
COMMENT ON COLUMN of_order.carid IS '卡编码';
COMMENT ON COLUMN of_order.cardname IS '卡编码';
COMMENT ON COLUMN of_order.cardnum IS '数量';
COMMENT ON COLUMN of_order.ordercash IS '订单金额(元)';
COMMENT ON COLUMN of_order.game_userid IS '手机号码';
COMMENT ON COLUMN of_order.game_state IS '0:充值中 1:成功 9:撤消,只能当状态为9时,商户才可以退款给用户';
COMMENT ON COLUMN of_order.reason IS '原因';
COMMENT ON COLUMN of_order.retcode IS '错误码';
COMMENT ON COLUMN of_order.create_time IS '注册时间';
COMMENT ON COLUMN of_order.update_time IS '更新时间';

CREATE TABLE "of_card"
(
  id bigserial,
  order_id bigint not null,  
  cardno varchar(32) default '',
  cardpws varchar(32) default '',
  expiretime varchar(32) default '', 
  ordercash numeric default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX idx_of_card ON order_of(order_id, cardno, expiretime);
ALTER TABLE of_card REPLICA IDENTITY FULL;
COMMENT ON TABLE of_card IS '欧飞卡密';
COMMENT ON COLUMN of_card.order_id IS '欧飞订单表id';
COMMENT ON COLUMN of_card.cardno IS '卡号';
COMMENT ON COLUMN of_card.cardpws IS '卡密';
COMMENT ON COLUMN of_card.expiretime IS '过期时间';
COMMENT ON COLUMN of_card.ordercash IS '订单金额(元)';
COMMENT ON COLUMN of_card.create_time IS '注册时间';
COMMENT ON COLUMN of_card.update_time IS '更新时间';


CREATE OR REPLACE FUNCTION fn_product_sku_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  total_sold int;
BEGIN
    -- 更新已售
    IF NEW.sold < 0 THEN
        RAISE EXCEPTION 'product_sku sold can not less 0 NEW:%', NEW;
    END IF;
    select sum(sold) into total_sold from product_sku where product_id=NEW.product_id;
    update product set sold=total_sold where id=NEW.product_id;

    RETURN NEW;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建平台资产明细触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_product_sku on product_sku;
CREATE TRIGGER tg_to_update_product_sku
after UPDATE of sold
ON product_sku  -- 指定触发表
FOR EACH ROW     -- 语句触发
EXECUTE PROCEDURE fn_product_sku_value();



CREATE OR REPLACE FUNCTION fn_product_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  sold numeric;
BEGIN
    -- 更新已售
    sold = NEW.sold;
    IF sold < 0 THEN
        RAISE EXCEPTION 'product sold can not less 0 NEW:%', NEW;
    END IF;

    RETURN NEW;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建平台资产明细触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_product on product;
CREATE TRIGGER tg_to_update_product
before UPDATE of sold
ON product  -- 指定触发表
FOR EACH ROW     -- 语句触发
EXECUTE PROCEDURE fn_product_value();