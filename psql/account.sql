/**
CREATE USER ybaccount WITH PASSWORD '123456';
CREATE DATABASE ybaccount with owner=ybaccount ENCODING='UTF8';
GRANT ALL PRIVILEGES ON DATABASE ybaccount to ybaccount;
ALTER DATABASE ybapi SET TIMEZONE='PRC';
\c ybaccount ybaccount;
**/


-- 创建array连接聚合
DROP AGGREGATE IF EXISTS anyarray_agg(anyarray);
CREATE AGGREGATE anyarray_agg(anyarray) (
  SFUNC = array_cat,
  STYPE = anyarray
);

CREATE FUNCTION unix_timestamp() RETURNS integer AS $$ 
SELECT (date_part('epoch',now()))::integer;   
$$ LANGUAGE SQL IMMUTABLE;

CREATE FUNCTION from_unixtime(int) RETURNS timestamp AS $$ 
SELECT to_timestamp($1)::timestamp; 
$$ LANGUAGE SQL IMMUTABLE;



CREATE TABLE "account"
(
  user_id bigserial,
  cc varchar(16) null,
  tel varchar(32) not null,
  "password" varchar(128) not null,
  status smallint DEFAULT '0',
  user_type smallint NOT NULL DEFAULT 0,
  platform varchar(20) NOT NULL,
  version varchar(20) NOT NULL default '',
  device_id varchar(64) NOT NULL,
  username varchar(128) default null, --用户名,可以重复.
  avatar varchar(256) default null, 
  birthday varchar(20) default '',
  zjpassword varchar(128) DEFAULT '', 
  cert_status smallint default -1,
  ip varchar(20) DEFAULT '',
  date varchar(20) DEFAULT '',
  country smallint default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(user_id)
);

CREATE UNIQUE INDEX account_cc_tel_uiq ON "account"(cc,tel);
CREATE INDEX account_create_time_idx ON "account"(create_time);
CREATE INDEX account_country_idx ON "account"(country, date, status, user_type);
ALTER TABLE account REPLICA IDENTITY FULL;
COMMENT ON TABLE account IS '用户表';
COMMENT ON COLUMN account.user_id IS '用户id';
COMMENT ON COLUMN account.cc IS '国家代码(手机)';
COMMENT ON COLUMN account.tel IS '手机号码';
COMMENT ON COLUMN account.password IS '密码';
COMMENT ON COLUMN account.status IS '用户状态(0:正常 1:冻结状态)';
COMMENT ON COLUMN account.user_type IS '用户类型(0:普通用户 1:商家用户)';
COMMENT ON COLUMN account.platform IS '注册平台';
COMMENT ON COLUMN account.version IS '注册时app版本';
COMMENT ON COLUMN account.device_id IS '设备id';
COMMENT ON COLUMN account.username IS '用户名';
COMMENT ON COLUMN account.avatar IS '头像URL';
COMMENT ON COLUMN account.zjpassword IS '资金密码';
COMMENT ON COLUMN account.cert_status IS '实名状态(-1:未实名 0:待实名 1:已实名 2:审核失败)';
COMMENT ON COLUMN account.ip IS '注册ip';
COMMENT ON COLUMN account.date IS '注册日期';
COMMENT ON COLUMN account.create_time IS '注册时间';
COMMENT ON COLUMN account.update_time IS '更新时间';
COMMENT ON COLUMN account.country IS '国家(0:国际版 1:国内版)';


select setval('account_user_id_seq', 51868);   --设置用户id初始值


CREATE TABLE "login_record"
(
  id bigserial,
  user_id bigint,
  ip varchar(20),
  country smallint DEFAULT 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX login_record_userid_create_time_idx ON "login_record"(user_id, create_time);
ALTER TABLE login_record REPLICA IDENTITY FULL;
COMMENT ON TABLE login_record IS '登录历史表';
COMMENT ON COLUMN login_record.user_id IS '用户id';
COMMENT ON COLUMN login_record.ip IS '登录IP';
COMMENT ON COLUMN login_record.create_time IS '注册时间';
COMMENT ON COLUMN login_record.update_time IS '更新时间';
COMMENT ON COLUMN login_record.country IS '国家(0:国际版 1:国内版)';



CREATE TABLE "cert"
(
  id bigserial,
  user_id bigint,
  card_country varchar(128) DEFAULT '',
  card_name varchar(128) DEFAULT '',
  card_id varchar(64) DEFAULT '',
  card_imgs jsonb default NULL,  
  status smallint default 0,
  country smallint default 0,
  reason varchar(512) default '',
  maner varchar(50) default '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX cert_uiq ON "cert"(user_id);
CREATE INDEX cert_idx ON "cert"(status, create_time);
ALTER TABLE cert REPLICA IDENTITY FULL;
COMMENT ON TABLE cert IS '实名认证表';
COMMENT ON COLUMN cert.user_id IS '用户id';
COMMENT ON COLUMN cert.card_country IS '国家';
COMMENT ON COLUMN cert.card_name IS '实名';
COMMENT ON COLUMN cert.card_id IS '身份证号';
COMMENT ON COLUMN cert.card_imgs IS '身份证件图片';
COMMENT ON COLUMN cert.status IS '认证状态(0:待审核 1:审核通过 2:审核不通过)';
COMMENT ON COLUMN cert.reason IS '原因';
COMMENT ON COLUMN cert.maner IS '审核人';
COMMENT ON COLUMN cert.create_time IS '注册时间';
COMMENT ON COLUMN cert.update_time IS '更新时间';
COMMENT ON COLUMN cert.country IS '国家(0:国际版 1:国内版)';

CREATE TABLE "imtoken"
(
  id bigserial,
  user_id bigint DEFAULT 0,
  imid varchar(32) DEFAULT '',
  token varchar(128) DEFAULT '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX imtoken_uidx ON imtoken(user_id);
ALTER TABLE imtoken REPLICA IDENTITY FULL;
COMMENT ON TABLE imtoken IS 'im token信息表';
COMMENT ON COLUMN imtoken.user_id IS '用户ID';
COMMENT ON COLUMN imtoken.imid IS 'imid';
COMMENT ON COLUMN imtoken.token IS 'im token';
COMMENT ON COLUMN imtoken.create_time IS '创建时间';
COMMENT ON COLUMN imtoken.update_time IS '更新时间';

CREATE TABLE "third_account"
(
  id bigserial,
  user_id bigint,
  third_name varchar(20) default 0,
  third_id bigint default 0,
  third_account jsonb default '{}',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX third_account_idx ON "third_account"(third_name, third_id);
ALTER TABLE third_account REPLICA IDENTITY FULL;
COMMENT ON TABLE third_account IS '第三方帐号关联';
COMMENT ON COLUMN third_account.user_id IS '用户id';
COMMENT ON COLUMN third_account.third_name IS '第三方标识';
COMMENT ON COLUMN third_account.third_id IS '第三方用户id';
COMMENT ON COLUMN third_account.third_account IS '第三方帐号信息';
COMMENT ON COLUMN third_account.create_time IS '注册时间';
COMMENT ON COLUMN third_account.update_time IS '更新时间';


