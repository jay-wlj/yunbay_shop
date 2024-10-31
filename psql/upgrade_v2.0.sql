\c ybapi ybapi

DROP TABLE IF EXISTS lotterys;
CREATE TABLE "lotterys"
(
  id bigserial,
  p_id bigint not null default 0,
  start_time int,
  end_time int,
  coin smallint default 0,
  price NUMERIC not null default 0,
  num smallint default 1,  
  stock smallint default 0,
  sold smallint default 0,
  reward_ybt numeric default 0,
  amount numeric default 0,
  status smallint default 0,
  hid smallint default 0,
  pertimes smallint default 1,
  product jsonb not null default '{}',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX idx_lotterys ON "lotterys" USING btree ("p_id", "start_time", "end_time");
COMMENT ON TABLE lotterys IS '抽奖活动表';
COMMENT ON COLUMN lotterys.p_id IS '抽奖商品id';
COMMENT ON COLUMN lotterys.start_time IS '抽奖开始时间';
COMMENT ON COLUMN lotterys.end_time IS '抽奖结束时间';
COMMENT ON COLUMN lotterys.coin IS '抽奖币种';
COMMENT ON COLUMN lotterys.price IS '抽奖总价';
COMMENT ON COLUMN lotterys.num IS '奖品数量';
COMMENT ON COLUMN lotterys.stock IS '总份数';
COMMENT ON COLUMN lotterys.sold IS '已售份数';
COMMENT ON COLUMN lotterys.amount IS '每次支付数量';
COMMENT ON COLUMN lotterys.reward_ybt IS '每次获得奖励ybt数量';
COMMENT ON COLUMN lotterys.status IS '当前活动状态(0:未开始 1:销售中 2:销售结束)';
COMMENT ON COLUMN lotterys.hid IS '是否隐藏';
COMMENT ON COLUMN lotterys.product IS '商品信息';
COMMENT ON COLUMN lotterys.pertimes IS '每人参与次数';
COMMENT ON COLUMN lotterys.create_time IS '创建时间';
COMMENT ON COLUMN lotterys.update_time IS '更新时间';

DROP TABLE IF EXISTS lotterys_record;
CREATE TABLE "lotterys_record"
(
  id bigserial,
  lotterys_id bigint not null default 0,
  user_id bigint default 0,
  --title varchar default '',
  --coin smallint default 0,
  amount numeric default 0,
  memo varchar(12) default '',
  hash varchar(72) default '',
  num_hash varchar(72) default '',
  status smallint default 0,
  order_status smallint default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX uidx_lotterys_record ON "lotterys_record" USING btree ("lotterys_id", "hash");
CREATE INDEX idx_lotterys_record ON "lotterys_record" USING btree ("user_id", "status");
COMMENT ON TABLE lotterys_record IS '抽奖记录表';
COMMENT ON COLUMN lotterys_record.lotterys_id IS '抽奖活动id';
COMMENT ON COLUMN lotterys_record.user_id IS '抽奖用户id';
COMMENT ON COLUMN lotterys_record.coin IS '支付币种';
COMMENT ON COLUMN lotterys_record.amount IS '支付数量';
COMMENT ON COLUMN lotterys_record.memo IS '支付唯一码';
COMMENT ON COLUMN lotterys_record.hash IS '交易hash';
COMMENT ON COLUMN lotterys_record.num_hash IS '数字hash';
COMMENT ON COLUMN lotterys_record.status IS '当前抽奖状态(-1:已返回 0:待开奖 1:未中奖 2:已中奖)';
COMMENT ON COLUMN lotterys_record.order_status IS '订单确认状态(0:未确认 1:已确认)';
COMMENT ON COLUMN lotterys_record.create_time IS '创建时间';
COMMENT ON COLUMN lotterys_record.update_time IS '更新时间';


\c ybasset ybasset


CREATE TABLE "transfer_pool"
(
  id bigserial,
  key varchar(32) not null default '',
  coin_type smallint DEFAULT 1,
  "from" bigint not null,
  "to" bigint not null,
  amount numeric not null default 0,
  status SMALLINT DEFAULT 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);

CREATE INDEX idx_transfer_pool ON transfer_pool(key, coin_type, "from", status);
ALTER TABLE transfer_pool REPLICA IDENTITY FULL;
COMMENT ON TABLE transfer_pool IS '转帐记录表';
COMMENT ON COLUMN transfer_pool.key IS '转帐key';
COMMENT ON COLUMN transfer_pool.coin_type IS '币种';
COMMENT ON COLUMN transfer_pool.amount IS '转帐数量';
COMMENT ON COLUMN transfer_pool.from IS '转帐发起方';
COMMENT ON COLUMN transfer_pool.to IS '转帐收款方';
COMMENT ON COLUMN transfer_pool.status IS '状态(-1: 退回 0:待转帐 1:成功)';
COMMENT ON COLUMN transfer_pool.create_time IS '注册时间';
COMMENT ON COLUMN transfer_pool.update_time IS '更新时间';


\c ybapi ybapi
insert into admin_controller (id, module, controller, name, create_time, update_time) values(10, 'backend', 'PublicOperation', '运营管理(通用))))', '1542104757', 1542104757);
insert into "admin_action" values (105, 10, 'LotterysList', '/backend/public-operation/lotterys-list', '积分抽奖任务列表', 1542104757, 1542104757);
insert into "admin_action" values (106, 10, 'LotterysDetail', '/backend/public-operation/lotterys-detail', '积分抽奖任务详情', 1542104757, 1542104757);
insert into "admin_action" values (107, 10, 'LotterysUpsert', '/backend/public-operation/lotterys-upsert', '修改积分抽奖任务', 1542104757, 1542104757);




CREATE TABLE "im_msgs"
(
  id bigserial,  
  type smallint default 0,
  msg jsonb not null default '{}',
  uids bigint[] default '{}',
  ok_uids bigint[] default '{}',
  status SMALLINT DEFAULT 0,
  ack boolean default false,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);

CREATE INDEX idx_im_msgs ON im_msgs( status);
ALTER TABLE im_msgs REPLICA IDENTITY FULL;
COMMENT ON TABLE im_msgs IS '下发消息记录表';
COMMENT ON COLUMN im_msgs.msg IS '消息内容';
COMMENT ON COLUMN im_msgs.uids IS '发送用户群体';
COMMENT ON COLUMN im_msgs.ok_uids IS '已成功发送用户';
COMMENT ON COLUMN im_msgs.status IS '状态(-1:失效 0:生效)';
COMMENT ON COLUMN im_msgs.ack IS '是否收到应答';
COMMENT ON COLUMN im_msgs.create_time IS '注册时间';
COMMENT ON COLUMN im_msgs.update_time IS '更新时间';


-- alter table currency_rate add column user_id bigint not null default 0;
-- drop index currency_rate_idx;
-- CREATE UNIQUE INDEX uidx_currency_rate ON "currency_rate"(key, user_id);



alter table product add column discount numeric(4,3) not null default 1;
