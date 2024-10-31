-- 用户奖励发放状态
CREATE TABLE "reward_record"
(
  id bigserial,
  type smallint default 0,
  release_type smallint default 0,
  fixdays int default 0,   
  invite_id bigint default 0,
  user_id bigint default 0,
  amount numeric default 0,
  lock boolean default false,
  reason varchar(50) default '',  
  maner varchar(64) default '',
  date varchar(20) default '',
  status smallint default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);

CREATE INDEX reward_record_idx ON "reward_record"(type, user_id, status, create_time);
ALTER TABLE reward_record REPLICA IDENTITY FULL;
COMMENT ON TABLE reward_record IS '用户奖励发放状态';
COMMENT ON COLUMN reward_record.id IS '赠送id';
COMMENT ON COLUMN reward_record.invite_id IS '邀请的用户id';
COMMENT ON COLUMN reward_record.user_id IS '用户id';
COMMENT ON COLUMN reward_record.type IS '币种类型(0:ybt 1:kt)';
COMMENT ON COLUMN reward_record.release_type IS '发放类型(0:空投, 1:活动奖励)';
COMMENT ON COLUMN reward_record.fixdays IS '定期天数';
COMMENT ON COLUMN reward_record.amount IS '数量';
COMMENT ON COLUMN reward_record.lock IS '空投奖励是否锁定状态(true为锁定，即不进行空投)';
COMMENT ON COLUMN reward_record.reason IS '说明';
COMMENT ON COLUMN reward_record.maner IS '操作人';
COMMENT ON COLUMN reward_record.date IS '日期';
COMMENT ON COLUMN reward_record.status IS '是否已发放(0:未发放 1:已发放)';
COMMENT ON COLUMN reward_record.create_time IS '注册时间';
COMMENT ON COLUMN reward_record.update_time IS '更新时间';


-- 用户kt分红每日明细
CREATE TABLE kt_bonus_detail(    
    id bigserial,
    user_id bigint,  
    total_ybt numeric default 0,
    normal_ybt numeric default 0,
    lock_ybt numeric default 0,  
    freeze_ybt numeric default 0,
    total_kt numeric default 0,
    normal_kt numeric default 0,
    lock_kt numeric default 0,    
    status smallint DEFAULT 0,    
    mining numeric default 0,
    air_unlock numeric default 0,
    project numeric default 0,
    bonus_ybt numeric default 0,    
    bonus_percent numeric default 0,
    kt_bonus numeric default 0,
    total_snet DOUBLE PRECISION DEFAULT 0,
    normal_snet DOUBLE PRECISION DEFAULT 0,
    lock_snet DOUBLE PRECISION DEFAULT 0,
    check_status smallint default 0,
    third_bonus smallint default 0,
    date varchar(20) default '',
    create_time int default 0,
    update_time int default 0,    
    primary key("id")
);
CREATE UNIQUE INDEX kt_bonus_detail_idx ON "kt_bonus_detail"(user_id, date);
CREATE INDEX kt_bonus_detail_user_idx ON "kt_bonus_detail"(check_status, create_time, update_time);
ALTER TABLE kt_bonus_detail REPLICA IDENTITY FULL;
COMMENT ON TABLE kt_bonus_detail IS '用户资产快照表';
COMMENT ON COLUMN kt_bonus_detail.user_id IS '用户id';
COMMENT ON COLUMN kt_bonus_detail.total_ybt IS '持有ybt总量';
COMMENT ON COLUMN kt_bonus_detail.total_kt IS '持有kt总量';
COMMENT ON COLUMN kt_bonus_detail.lock_ybt IS '锁定(冻结)ybt总量';
COMMENT ON COLUMN kt_bonus_detail.lock_kt IS '锁定(冻结)kt总量';
COMMENT ON COLUMN kt_bonus_detail.freeze_ybt IS '空投锁定的ybt总量';
COMMENT ON COLUMN kt_bonus_detail.normal_ybt IS '可用ybt总量';
COMMENT ON COLUMN kt_bonus_detail.normal_kt IS '可用kt总量';
COMMENT ON COLUMN kt_bonus_detail.status IS '帐户冻结类型(0:未冻结 1:已冻结)';
COMMENT ON COLUMN kt_bonus_detail.mining IS '挖矿释放(ybt)';
COMMENT ON COLUMN kt_bonus_detail.project IS '项目奖励释放(ybt)';
COMMENT ON COLUMN kt_bonus_detail.air_unlock IS '空投释放(ybt)';
COMMENT ON COLUMN kt_bonus_detail.bonus_ybt IS '可分红的ybt';
COMMENT ON COLUMN kt_bonus_detail.bonus_percent IS '可分红占比';
COMMENT ON COLUMN kt_bonus_detail.kt_bonus IS '收益金';
COMMENT ON COLUMN kt_bonus_detail.check_status IS 'kt收益金是否发放(0:否 1:是)';
COMMENT ON COLUMN kt_bonus_detail.third_bonus IS '是否第三方分红平台(0:否 1:是)';
COMMENT ON COLUMN kt_bonus_detail.date IS '日期';
COMMENT ON COLUMN kt_bonus_detail.create_time IS '注册时间';
COMMENT ON COLUMN kt_bonus_detail.update_time IS '更新时间';
COMMENT ON COLUMN kt_bonus_detail.total_snet IS '持有的snet';
COMMENT ON COLUMN kt_bonus_detail.normal_snet IS '可用的snet';
COMMENT ON COLUMN kt_bonus_detail.lock_snet IS '锁定的snet';

-- 用户挖矿快每日明细
CREATE TABLE "ybt_unlock_detail"
(
  id bigserial,
  user_id bigint,  
  mining numeric default 0,
  consume numeric default 0,
  sale numeric default 0,
  invite numeric default 0,
  activity numeric default 0,  
  air_drop numeric default 0,  
  air_unlock numeric default 0,
  project numeric default 0,
  total_unlock numeric default 0,
  ybt_percent float default 0,
  rebat numeric default 0,
  rebat_percent float default 0,
  sale_rebat numeric default 0,
  sale_percent float default 0, 
  check_status smallint DEFAULT 0,    
  date varchar(20) default '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX ybt_unlock_detail_user_idx ON "ybt_unlock_detail"(user_id, date);
CREATE INDEX ybt_unlock_detail_idx ON "ybt_unlock_detail"(check_status, create_time, update_time);
ALTER TABLE ybt_unlock_detail REPLICA IDENTITY FULL;
COMMENT ON TABLE ybt_unlock_detail IS 'ybt解锁详细信息';
COMMENT ON COLUMN ybt_unlock_detail.user_id IS '用户id';
COMMENT ON COLUMN ybt_unlock_detail.mining IS '挖矿奖励(ybt)';
COMMENT ON COLUMN ybt_unlock_detail.consume IS '消费奖励(ybt)';
COMMENT ON COLUMN ybt_unlock_detail.sale IS '商家奖励(ybt)';
COMMENT ON COLUMN ybt_unlock_detail.invite IS '邀请奖励(ybt)';
COMMENT ON COLUMN ybt_unlock_detail.activity IS '活动奖励(ybt)';
COMMENT ON COLUMN ybt_unlock_detail.air_drop IS '空投奖励(ybt)';
COMMENT ON COLUMN ybt_unlock_detail.air_unlock IS '空投释放(ybt)';
COMMENT ON COLUMN ybt_unlock_detail.project IS '项目释放(ybt)';
COMMENT ON COLUMN ybt_unlock_detail.total_unlock IS '总释放(ybt)';
COMMENT ON COLUMN ybt_unlock_detail.ybt_percent IS 'ybt占比';
COMMENT ON COLUMN ybt_unlock_detail.rebat IS '贡献值(kt)';
COMMENT ON COLUMN ybt_unlock_detail.rebat_percent IS '贡献贡献占比';
COMMENT ON COLUMN ybt_unlock_detail.sale_rebat IS '商家贡献值(kt)';
COMMENT ON COLUMN ybt_unlock_detail.sale_percent IS '商家贡献贡献占比';
COMMENT ON COLUMN ybt_unlock_detail.check_status IS '发放状态(0:未发放 1:已发放)';
COMMENT ON COLUMN ybt_unlock_detail.date IS '日期';
COMMENT ON COLUMN ybt_unlock_detail.create_time IS '注册时间';
COMMENT ON COLUMN ybt_unlock_detail.update_time IS '更新时间';

CREATE TABLE "ybt"
(
  id bigserial,
  reward numeric default 0,
  project numeric default 0,
  minepool numeric default 0,  
  normal_reward numeric default 0,  
  lock_reward numeric default 0,
  lock_project numeric default 0,
  lock_minepool numeric default 0,
  unlock_reward numeric default 0,
  unlock_project numeric default 0,
  unlock_minepool numeric default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX ybt_idx ON "ybt"(create_time);
ALTER TABLE ybt REPLICA IDENTITY FULL;
COMMENT ON TABLE ybt IS 'ybt发行总表';
COMMENT ON COLUMN ybt.reward IS '空投奖励';
COMMENT ON COLUMN ybt.project IS '项目方';
COMMENT ON COLUMN ybt.minepool IS '矿池';
COMMENT ON COLUMN ybt.normal_reward IS '可用的空投奖励';
COMMENT ON COLUMN ybt.lock_reward IS '锁定的奖励';
COMMENT ON COLUMN ybt.lock_project IS '锁定的项目方';
COMMENT ON COLUMN ybt.lock_minepool IS '锁定的矿池';
COMMENT ON COLUMN ybt.unlock_reward IS '已解锁用户奖励';
COMMENT ON COLUMN ybt.unlock_project IS '已解锁项目方';
COMMENT ON COLUMN ybt.unlock_minepool IS '已解锁矿池量';
COMMENT ON COLUMN ybt.create_time IS '注册时间';
COMMENT ON COLUMN ybt.update_time IS '更新时间';

-- 预发行代币释放
insert into ybt(reward,project,minepool,normal_reward,lock_reward,lock_project,lock_minepool,create_time, update_time) values(100000000, 400000000, 500000000, 100000000, 100000000, 400000000, 500000000, unix_timestamp(), unix_timestamp());


-- ybt奖励当日记录
CREATE TABLE "ybt_day_flow"
(
  id bigserial,
  reward numeric default 0,  
  unlock_reward numeric default 0,
  unlock_project numeric default 0,
  unlock_minepool numeric default 0,    
  date varchar(20) default '',
  maner varchar(20) default '',
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE UNIQUE INDEX ybt_day_flow_idx ON "ybt_day_flow"(date);
ALTER TABLE ybt_day_flow REPLICA IDENTITY FULL;
COMMENT ON TABLE ybt_day_flow IS 'ybt释放记录';
COMMENT ON COLUMN ybt_day_flow.reward IS '空投奖励';
COMMENT ON COLUMN ybt_day_flow.unlock_reward IS '释放空投数量';
COMMENT ON COLUMN ybt_day_flow.unlock_project IS '释放项目方量';
COMMENT ON COLUMN ybt_day_flow.unlock_minepool IS '释放矿池量';
COMMENT ON COLUMN ybt_day_flow.date IS '日期';
COMMENT ON COLUMN ybt_day_flow.maner IS '审核人';
COMMENT ON COLUMN ybt_day_flow.create_time IS '注册时间';
COMMENT ON COLUMN ybt_day_flow.update_time IS '更新时间';


-- ybt奖励记录
CREATE TABLE "ybt_flow"
(
  id bigserial,
  type smallint default 0,
  user_id bigint default 0,
  amount numeric default 0,
  user_asset_id bigint default 0,
  date varchar(20) default '',
  maner varchar(64) default 0,
--status smallint default 0,
  create_time int default 0,
  update_time int default 0,
  primary key(id)
);
CREATE INDEX ybt_flow_user_idx ON "ybt_flow"(type, user_id, date);
ALTER TABLE ybt_flow REPLICA IDENTITY FULL;
COMMENT ON TABLE ybt_flow IS 'ybt奖励用户流水记录';
COMMENT ON COLUMN ybt_flow.id IS 'ybt奖励id';
COMMENT ON COLUMN ybt_flow.type IS 'ybt奖励类型(0:挖矿奖励 1:空投奖励 2:活动奖励 3:项目释放)';
COMMENT ON COLUMN ybt_flow.user_id IS '奖励的用户id';
COMMENT ON COLUMN ybt_flow.amount IS '奖励的ybt数量';
COMMENT ON COLUMN ybt_flow.user_asset_id IS '关联已奖励的用户资产流水id';
COMMENT ON COLUMN ybt_flow.date IS '奖励日期';
-- COMMENT ON COLUMN reward_ybt_flow.status IS '状态(0:待奖励 1:已奖励)';
COMMENT ON COLUMN ybt_flow.create_time IS '注册时间';
COMMENT ON COLUMN ybt_flow.update_time IS '更新时间';





CREATE TABLE "bonus_ybt"
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
CREATE UNIQUE INDEX bonus_ybt_uidx ON "bonus_ybt"(user_id, date);
ALTER TABLE bonus_ybt REPLICA IDENTITY FULL;
COMMENT ON TABLE bonus_ybt IS '用户每天ybt奖励明细表';
COMMENT ON COLUMN bonus_ybt.user_id IS '用户id';
COMMENT ON COLUMN bonus_ybt.infos IS '类型信息';
COMMENT ON COLUMN bonus_ybt.total_ybt IS '累积奖励ybt';
COMMENT ON COLUMN bonus_ybt.date IS '日期';
COMMENT ON COLUMN bonus_ybt.create_time IS '注册时间';
COMMENT ON COLUMN bonus_ybt.update_time IS '更新时间';


-- 对ybt发放进行按月分表存储
CREATE OR REPLACE FUNCTION bonus_ybt_partition()
RETURNS TRIGGER AS $$
DECLARE date_text TEXT;
DECLARE insert_statement TEXT;
DECLARE table_name TEXT;
BEGIN	
	SELECT to_char(NEW.date::timestamp, 'YYYY_MM') INTO date_text;
    table_name = 'bonus_ybt_' || date_text;
	insert_statement := 'INSERT INTO ' || table_name ||' VALUES ($1.*)';
	EXECUTE insert_statement USING NEW;
	RETURN NEW;     -- 此处需要返回NEW  不能返回NULL
	EXCEPTION
	WHEN UNDEFINED_TABLE
	THEN
		EXECUTE
			'CREATE TABLE IF NOT EXISTS ' || table_name || '(CHECK (to_char(date::timestamp, ''YYYY_MM'')=''' || date_text
			|| ''')) INHERITS (bonus_ybt)';
		RAISE NOTICE 'CREATE NON-EXISTANT TABLE %', table_name;
		EXECUTE
			'CREATE UNIQUE INDEX ' || table_name || '_idx ON "'|| table_name || '"(user_id, date)';
		EXECUTE insert_statement USING NEW;
    RETURN NEW;     -- 此处需要返回NEW  不能返回NULL
END;
$$
LANGUAGE plpgsql;


CREATE TRIGGER insert_bonus_ybt_partition
BEFORE INSERT ON bonus_ybt
FOR EACH ROW EXECUTE PROCEDURE bonus_ybt_partition();


CREATE OR REPLACE FUNCTION bonus_ybt_after_partition()
RETURNS TRIGGER AS $$
DECLARE r bonus_ybt%rowtype;
BEGIN
	delete from only bonus_ybt where id=NEW.id returning * into r;      -- 需要后触发删除主表插入相同的记录
  return r;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER insert_after_bonus_ybt_partition
AFTER INSERT ON bonus_ybt 
FOR EACH ROW EXECUTE PROCEDURE bonus_ybt_after_partition();


insert into bonus_ybt select * from bonus_ybt_detail;
alter table bonus_ybt_detail rename to bonus_ybt_detail_bak;		-- 先备份 以后再删除
alter table bonus_ybt rename to bonus_ybt_detail;


CREATE TABLE "bonus_kt"
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
CREATE UNIQUE INDEX bonus_kt_idx ON "bonus_kt"(user_id, date);
ALTER TABLE bonus_kt REPLICA IDENTITY FULL;
COMMENT ON TABLE bonus_kt IS '用户每天kt收益明细表';
COMMENT ON COLUMN bonus_kt.user_id IS '用户id';
COMMENT ON COLUMN bonus_kt.ybt IS '可分红ybt';
COMMENT ON COLUMN bonus_kt.kt IS 'kt收益金';
COMMENT ON COLUMN bonus_kt.date IS '日期';
COMMENT ON COLUMN bonus_kt.create_time IS '注册时间';
COMMENT ON COLUMN bonus_kt.update_time IS '更新时间';

-- 对kt发放进行按月分表存储
CREATE OR REPLACE FUNCTION bonus_kt_partition()
RETURNS TRIGGER AS $$
DECLARE date_text TEXT;
DECLARE insert_statement TEXT;
DECLARE table_name TEXT;
BEGIN	
	SELECT to_char(NEW.date::timestamp, 'YYYY_MM') INTO date_text;
    table_name = 'bonus_kt_' || date_text;
	insert_statement := 'INSERT INTO ' || table_name ||' VALUES ($1.*)';
	EXECUTE insert_statement USING NEW;
	RETURN NEW;     -- 此处需要返回NEW  不能返回NULL
	EXCEPTION
	WHEN UNDEFINED_TABLE
	THEN
		EXECUTE
			'CREATE TABLE IF NOT EXISTS ' || table_name || '(CHECK (to_char(date::timestamp, ''YYYY_MM'')=''' || date_text
			|| ''')) INHERITS (bonus_kt)';
		RAISE NOTICE 'CREATE NON-EXISTANT TABLE %', table_name;
		EXECUTE
			'CREATE UNIQUE INDEX ' || table_name || '_idx ON "'|| table_name || '"(user_id, date)';
		EXECUTE insert_statement USING NEW;
    RETURN NEW;     -- 此处需要返回NEW  不能返回NULL
END;
$$
LANGUAGE plpgsql;

drop trigger insert_bonus_kt_partition on bonus_kt;
CREATE TRIGGER insert_bonus_kt_partition
BEFORE INSERT ON bonus_kt
FOR EACH ROW EXECUTE PROCEDURE bonus_kt_partition();


CREATE OR REPLACE FUNCTION bonus_kt_after_partition()
RETURNS TRIGGER AS $$
DECLARE r bonus_kt%rowtype;
BEGIN
	--delete from only bonus_kt where id=NEW.id returning * into r;      -- 需要后触发删除主表插入相同的记录
  delete from only bonus_kt;      -- 需要后触发删除主表插入相同的记录
  return NULL;
END;
$$
LANGUAGE plpgsql;

drop trigger insert_after_bonus_kt_partition on bonus_kt;
CREATE TRIGGER insert_after_bonus_kt_partition
AFTER INSERT ON bonus_kt 
--FOR EACH ROW EXECUTE PROCEDURE bonus_kt_after_partition();
FOR EACH STATEMENT EXECUTE PROCEDURE bonus_kt_after_partition();


insert into bonus_kt select * from bonus_kt_detail;