\c ybasset ybasset


-- 创建用户资产触发器
-- 创建用户资产触发器函数 保证用户资产不出现负数
CREATE OR REPLACE FUNCTION fn_user_asset_type_check_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  fnetgive numeric;
BEGIN
    -- 用户snet资产校验
    fnetgive = -0.000000001;
    IF NEW.total_amount < 0 THEN
      IF NEW.total_amount >= fnetgive THEN
          NEW.total_amount := 0;
      ELSE
          RAISE EXCEPTION '% total_amount cannot be negtive %', NEW.type, NEW.total_amount;
      END IF;
    END IF;

    IF NEW.lock_amount < 0 THEN
      IF NEW.lock_amount >= fnetgive THEN
          NEW.lock_amount := 0;
      ELSE
          RAISE EXCEPTION '% lock_amount cannot be negtive % ', NEW.type, NEW.lock_amount;
      END IF;
    END IF;

    IF NEW.normal_amount < 0 THEN
      IF NEW.normal_amount >= fnetgive THEN
          NEW.normal_amount := 0;
      ELSE
          RAISE EXCEPTION '% normal_amount cannot be negtive % ', NEW.type, NEW.normal_amount;
      END IF;
    END IF;

    IF NEW.freeze_amount < 0 THEN
      IF NEW.freeze_amount >= fnetgive THEN
          NEW.freeze_amount := 0;
      ELSE
          RAISE EXCEPTION '% freeze_amount cannot be negtive % ', NEW.type, NEW.freeze_amount;
      END IF;
    END IF;
    RETURN NEW; 
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建用户资产明细触发器，注意要为后触发，行级触发器

CREATE TRIGGER tg_to_before_update_user_total_asset
before INSERT OR UPDATE of total_amount,lock_amount
ON user_asset_type  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_user_asset_type_check_value();


CREATE OR REPLACE FUNCTION fn_user_asset_type_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  _normal_amount numeric;
  fnetgive numeric;
BEGIN
    fnetgive = -0.00000001;
    if NEW.type = 0 then 
      _normal_amount := NEW.total_amount-NEW.lock_amount-NEW.freeze_amount;
    else 
      _normal_amount := NEW.total_amount-NEW.lock_amount;
    end if;

    IF _normal_amount < 0 THEN   -- 限制锁定的ybt不能大于当前总ybt
      IF _normal_amount >= fnetgive THEN
        if NEW.type <> 0 then
          update user_asset_type set lock_amount=NEW.total_amount where id=NEW.id;   -- 更新lock_snet后会触发前后触发器
          RETURN null;
        end if;
        _normal_amount = 0;
      ELSE 
        RAISE EXCEPTION 'user_asset amount normal_amount:% NEW:%', _normal_amount, NEW;
      END IF;
    END IF;
    
    update user_asset_type set normal_amount=_normal_amount where id=NEW.id;
    

    -- 更新user_asset表中的部分资产
    if NEW.type = 0 THEN
      -- RAISE NOTICE '%',NEW;
      insert into user_asset(user_id, total_ybt, normal_ybt, lock_ybt, freeze_ybt) values(NEW.user_id, NEW.total_amount, _normal_amount, NEW.lock_amount, NEW.freeze_amount) on conflict(user_id) do update set total_ybt=NEW.total_amount, normal_ybt=_normal_amount, lock_ybt=NEW.lock_amount, freeze_ybt=NEW.freeze_amount;
    elsif NEW.type = 1 THEN   
      insert into user_asset(user_id, total_kt, normal_kt, lock_kt) values(NEW.user_id, NEW.total_amount, _normal_amount, NEW.lock_amount) on conflict(user_id) do update set total_kt=NEW.total_amount, normal_kt=_normal_amount, lock_kt=NEW.lock_amount;
    elsif NEW.type = 3 THEN
      insert into user_asset(user_id, total_snet, normal_snet, lock_snet) values(NEW.user_id, NEW.total_amount, _normal_amount, NEW.lock_amount) on conflict(user_id) do update set total_snet=NEW.total_amount, normal_snet=_normal_amount, lock_snet=NEW.lock_amount;
    end if;
    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建用户资产明细触发器，注意要为后触发，行级触发器
CREATE TRIGGER tg_to_update_user_asset_amount
after INSERT OR UPDATE of total_amount,lock_amount
ON user_asset_type  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_user_asset_type_value();


CREATE OR REPLACE FUNCTION fn_user_asset_detail_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  _amount numeric;
  obj record;
  UserId bigint;
  _type int;
  _now int;
BEGIN
    -- 自动更新用户资产表
    _now = unix_timestamp();
    _amount = 0;
    IF (TG_OP = 'DELETE') THEN    --删除记录只有OLD，新增和更新是NEW
      UserId = OLD.user_id;
      _type = OLD.type;
    ELSE
      UserId = NEW.user_id;
      _type = NEW.type;
    END IF;

    select sum(amount) as amount into obj from user_asset_detail where user_id=UserId and "type"=_type;
    if obj is not null then
      _amount = obj.amount;
    end if;      
    insert into user_asset_type(user_id, type, "total_amount", create_time, update_time) values(UserId, _type, _amount, _now, _now) on conflict(user_id, type) do update set total_amount=_amount, update_time=_now;

    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建用户资产明细触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_user_asset_detail on user_asset_detail;
CREATE TRIGGER tg_to_update_user_asset_detail
after INSERT OR DELETE OR UPDATE 
ON user_asset_detail  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_user_asset_detail_value();



CREATE OR REPLACE FUNCTION fn_asset_lock_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  lock_obj record;
  _lock_amount numeric;
  _lock_freeze numeric;
  obj record;
  _type int;
  UserId bigint;
  _now int;
BEGIN
    -- 自动更新用户资产表
    _now = unix_timestamp();
    IF (TG_OP = 'DELETE') THEN    --删除记录只有OLD，新增和更新是NEW
      UserId = OLD.user_id;
      _type = OLD.type;
    ELSE
      UserId = NEW.user_id;
      _type = NEW.type;
    END IF;

    _lock_amount = 0;
    _lock_freeze = 0;
    select sum(lock_amount) as lock_amount into obj from asset_lock where user_id=UserId and "type"=_type;
    if obj is not null then
        _lock_amount = obj.lock_amount;
    end if;
        
    if _type = 0 then -- ybt
      if _type = 0 and obj.lock_amount > 0 THEN        
        select sum(lock_amount) as amount into obj from asset_lock where type=_type and user_id=UserId and lock_type = 0;
        if obj is not null THEN
          _lock_freeze = obj.amount;
        end if;
      end if;
      update user_asset_type set lock_amount=_lock_amount-_lock_freeze, freeze_amount=_lock_freeze, update_time=_now where user_id=UserId and type=_type;    -- 更新ybt的总冻结和总锁定资产
    else   -- kt      
      update user_asset_type set lock_amount=_lock_amount, update_time=_now where user_id=UserId and type=_type;       -- 更新总冻结资产
    end if;

    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建用户资产明细触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_asset_lock on asset_lock;
CREATE TRIGGER tg_to_update_asset_lock
after INSERT OR DELETE OR UPDATE 
ON asset_lock  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_asset_lock_value();




-- 创建用户资产触发器
-- 创建用户资产触发器函数 保证用户资产不出现负数
CREATE OR REPLACE FUNCTION fn_user_asset_check_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  fnetgive numeric;
BEGIN
    --dt = current_date;
    -- 用户资产校验
    fnetgive = -0.000000001;
    IF NEW.total_ybt < 0 THEN
      IF NEW.total_ybt >= fnetgive THEN
          NEW.total_ybt := 0;
      ELSE
          RAISE EXCEPTION 'total_ybt cannot be negtive %', NEW.total_ybt;
      END IF;
    END IF;

    IF NEW.lock_ybt < 0 THEN
      IF NEW.lock_ybt >= fnetgive THEN
          NEW.lock_ybt := 0;
      ELSE
          RAISE EXCEPTION 'lock_ybt cannot be negtive %', NEW.lock_ybt;
      END IF;
    END IF;

    IF NEW.freeze_ybt < 0 THEN
      IF NEW.freeze_ybt >= fnetgive THEN
          NEW.freeze_ybt := 0;
      ELSE
          RAISE EXCEPTION 'freeze_ybt cannot be negtive %', NEW.freeze_ybt;
      END IF;
    END IF;

    IF NEW.lock_kt < 0 THEN
      IF NEW.lock_kt  >= fnetgive THEN
          NEW.lock_kt := 0;
      ELSE
          RAISE EXCEPTION 'lock_kt cannot be negtive %', NEW.lock_kt;
      END IF;
    END IF;
    
    IF NEW.total_kt < 0 THEN
      IF NEW.total_kt >= fnetgive THEN
          NEW.total_kt := 0;
      ELSE
          RAISE EXCEPTION 'total_kt cannot be negtive %', NEW.total_kt;
      END IF;
    END IF;


    if NEW.normal_ybt < 0 THEN
      IF NEW.normal_ybt >= fnetgive THEN
          NEW.normal_ybt := 0;
      ELSE
          RAISE EXCEPTION 'normal_ybt cannot be negtive %', NEW.normal_ybt;
      END IF;
    END IF;

    if NEW.normal_kt < 0 THEN
      IF NEW.normal_kt >= fnetgive THEN
          NEW.normal_kt := 0;
      ELSE
          RAISE EXCEPTION 'normal_kt cannot be negtive %', NEW.normal_kt;
      END IF;
    END IF;
    RETURN NEW;    
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建用户资产明细触发器，注意要为后触发，行级触发器
CREATE TRIGGER tg_to_before_update_user_asset
before INSERT OR UPDATE of total_ybt,total_kt,lock_ybt,lock_kt,freeze_ybt
ON user_asset  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_user_asset_check_value();


-- 创建用户资产触发器
-- 创建用户资产触发器函数 保证用户资产不出现负数
CREATE OR REPLACE FUNCTION fn_user_asset_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  _normal_ybt numeric;
  _normal_kt numeric;
  fnetgive numeric;
BEGIN
    fnetgive = -0.00000001;
    --dt = current_date;
    -- 用户资产校验
    -- RAISE EXCEPTION 'user_asset amount less than 0 (%), NEW:%', NEW, NEW;
    -- IF NEW.total_ybt<fnetgive or NEW.total_kt<fnetgive or NEW.lock_ybt<fnetgive or NEW.lock_kt<fnetgive or round((NEW.lock_ybt-NEW.total_ybt)::numeric, 12)>0.000000001 or round((NEW.lock_kt-NEW.total_kt)::numeric, 12)>0.000000001 THEN
    --   --RAISE EXCEPTION 'user_asset amount less than 0 %', NEW;
    --   RAISE EXCEPTION 'user_asset amount less than 0 (%)', NEW;
    -- END IF;

    --RAISE INFO 'fn_user_asset_value start';
    _normal_ybt = NEW.total_ybt-NEW.lock_ybt-NEW.freeze_ybt;
    _normal_kt = NEW.total_kt - NEW.lock_kt;
    
    IF _normal_ybt < 0 THEN   -- 限制锁定的ybt不能大于当前总ybt
      IF _normal_ybt >= fnetgive THEN
        update user_asset set lock_ybt=NEW.total_ybt where id=NEW.id;   -- 更新lock_ybt后会触发前后触发器
        RETURN null;
      ELSE 
        RAISE EXCEPTION 'user_asset amount normal_ybt:% NEW:%', _normal_ybt, NEW;
      END IF;
    END IF;
    
    IF _normal_kt < 0 THEN    -- 限制锁定的kt不能大于当前总kt
      IF _normal_kt >= fnetgive THEN
        update user_asset set lock_kt=NEW.total_kt where id=NEW.id;      -- 更新lock_ybt后会触发前后触发器
        RETURN null;
      ELSE 
        RAISE EXCEPTION 'user_asset amount normal_kt:%  NEW:%', _normal_kt, NEW;
      END IF;
    END IF;

    --RAISE INFO 'fn_user_asset_value end';

    update user_asset set normal_ybt=_normal_ybt, normal_kt=_normal_kt where id=NEW.id;   -- 不对normal资产触发
    
    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建用户资产明细触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_user_asset on user_asset;
CREATE TRIGGER tg_to_update_user_asset
after INSERT OR UPDATE of total_ybt,total_kt,lock_ybt,lock_kt
ON user_asset  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_user_asset_value();



-- 创建用户snet资产后触发器
-- 创建用户资产触发器函数 保证用户资产不出现负数
CREATE OR REPLACE FUNCTION fn_user_snet_asset_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  _normal_amount numeric;
  fnetgive numeric;
BEGIN
    fnetgive = -0.000000001;
    _normal_amount := NEW.total_snet-NEW.lock_snet;
    IF _normal_amount < 0 THEN   -- 限制锁定的ybt不能大于当前总ybt
      IF _normal_amount >= fnetgive THEN
        update user_asset set lock_snet=NEW.total_snet where id=NEW.id;   -- 更新lock_snet后会触发前后触发器
        RETURN null;
      ELSE 
        RAISE EXCEPTION 'user_asset amount normal_snet:% NEW:%', _normal_amount, NEW;
      END IF;
    END IF;

    update user_asset set normal_snet=_normal_amount where id=NEW.id;
    
    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建用户资产明细触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_user_snet_asset on user_asset;
CREATE TRIGGER tg_to_update_user_snet_asset
after INSERT OR UPDATE of total_snet,lock_snet
ON user_asset  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_user_snet_asset_value();



-- 创建用户资产触发器
-- 创建用户资产触发器函数 保证用户资产不出现负数
CREATE OR REPLACE FUNCTION fn_user_snet_asset_check_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  fnetgive numeric;
BEGIN
    -- 用户snet资产校验
    fnetgive = -0.000000001;
    IF NEW.total_snet < 0 THEN
      IF NEW.total_snet >= fnetgive THEN
          NEW.total_snet := 0;
      ELSE
          RAISE EXCEPTION 'total_snet cannot be negtive %', NEW.total_snet;
      END IF;
    END IF;

    IF NEW.lock_snet < 0 THEN
      IF NEW.lock_snet >= fnetgive THEN
          NEW.lock_snet := 0;
      ELSE
          RAISE EXCEPTION 'lock_snet cannot be negtive % ', NEW.lock_snet;
      END IF;
    END IF;

    IF NEW.normal_snet < 0 THEN
      IF NEW.normal_snet >= fnetgive THEN
          NEW.normal_snet := 0;
      ELSE
          RAISE EXCEPTION 'normal_snet cannot be negtive % ', NEW.normal_snet;
      END IF;
    END IF;

    RETURN NEW; 
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建用户资产明细触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_user_asset on user_asset;
CREATE TRIGGER tg_to_update_user_asset
after INSERT OR UPDATE of total_ybt,total_kt,lock_ybt,lock_kt
ON user_asset  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_user_asset_value();





CREATE OR REPLACE FUNCTION fn_voucher_record_amount_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  obj record;
	_amount numeric;
	_now bigint;
BEGIN
    -- 支付状态信息
    _now = NEW.update_time;   -- 用unix_timestamp()会有问题 可能与上次调用的值相同
    select sum(amount) into _amount from voucher_record where voucher_id=NEW.voucher_id;
		update voucher set amount=_amount, update_time=_now where id=NEW.voucher_id;

    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

drop trigger tg_to_update_voucher_record on voucher_record;
CREATE TRIGGER tg_to_update_voucher_record
after INSERT OR UPDATE of amount    -- 后触发
--BEFORE UPDATE 
ON voucher_record  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_voucher_record_amount_value();



CREATE OR REPLACE FUNCTION fn_voucher_amount_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  fnetgive numeric;
BEGIN
    fnetgive = -0.000000001;
    IF NEW.amount < 0 THEN
      IF NEW.amount >= fnetgive THEN
          NEW.amount := 0;
      ELSE
          RAISE EXCEPTION 'amount cannot be negtive %', NEW.amount;
      END IF;
    END IF;

    RETURN NEW; 
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建用户资产明细触发器，注意要为后触发，行级触发器
drop trigger tg_to_before_update_voucher_amount on voucher;
CREATE TRIGGER tg_to_before_update_voucher_amount
before INSERT OR UPDATE of amount
ON voucher  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_voucher_amount_value();


-- 创建ybt触发器
-- 创建ybt触发器函数 
CREATE OR REPLACE FUNCTION fn_ybt_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
    obj record;
BEGIN
    -- ybt解锁
    IF NEW.normal_reward<0 or round((NEW.normal_reward-NEW.reward)::numeric,12)>0.00000000001 or NEW.unlock_project<0 or round((NEW.unlock_project-NEW.project)::numeric,12)>0.00000000001 or NEW.unlock_minepool<0 or round((new.unlock_minepool-NEW.minepool)::numeric,12)>0.00000000001 THEN
      --RAISE EXCEPTION 'user_asset amount less than 0 %', NEW;
      RAISE EXCEPTION 'ybt amount less than 0 (%)', NEW;
    END IF;

    -- 冻结的空投ybt=已空投的ybt-已解锁的ybt    
    update ybt set lock_reward=round((NEW.reward-NEW.normal_reward-NEW.unlock_reward)::numeric,12), lock_project=NEW.project-NEW.unlock_project, lock_minepool=NEW.minepool-NEW.unlock_minepool;
    
    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建ybt解锁触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_ybt on ybt;
CREATE TRIGGER tg_to_update_ybt
after INSERT OR UPDATE of normal_reward,unlock_reward,unlock_project,unlock_minepool
ON ybt  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_ybt_value();




-- 创建ybt每日总流水触发器
-- 创建ybt每日总流水触发器函数 
CREATE OR REPLACE FUNCTION fn_ybt_day_flow_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
    obj record;
    _now int;
BEGIN
    -- ybt解锁
    select sum(reward) as reward, sum(unlock_reward) as unlock_reward, sum(unlock_project) as unlock_project, sum(unlock_minepool) as unlock_minepool into obj from ybt_day_flow;
    if obj is not null then
        _now = unix_timestamp();
        update ybt set normal_reward=ybt.reward-obj.reward, unlock_reward=obj.unlock_reward, unlock_project=obj.unlock_project, unlock_minepool=obj.unlock_minepool, update_time=_now;
    end if;   
    
    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建ybt每日总流水触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_ybt_day_flow on ybt_day_flow;
CREATE TRIGGER tg_to_update_ybt_day_flow
after INSERT OR DELETE OR UPDATE 
ON ybt_day_flow  -- 指定触发表
FOR EACH STATEMENT     -- 语句触发  
EXECUTE PROCEDURE fn_ybt_day_flow_value();




-- 创建ybt解锁触发器
-- 创建ybt解锁触发器函数 
CREATE OR REPLACE FUNCTION fn_ybt_flow_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
    obj record;
    _type int;
    _date varchar(20);
    _now int;
BEGIN
     IF (TG_OP = 'DELETE') THEN
        _type = OLD.type;
        _date = OLD.date;
     ELSE
        _type = NEW.type;
        _date = NEW.date;
     END IF;

     _now = unix_timestamp();
    -- ybt解锁
    if _type = 0 or _type = 1 then    -- 空投及活动赠送奖励都算空投奖励
            select sum(amount) as amount into obj from ybt_flow where type in (0, 1) and date=_date;
            if obj is not null then
                insert into ybt_day_flow(reward, date, create_time, update_time) values(obj.amount, _date, _now, _now) on conflict(date) do update
                set reward = obj.amount, update_time=_now;     
            end if;
    else 
        select sum(amount) as amount into obj from ybt_flow where type=_type and date=_date;
        if obj is null THEN
            return null;
        end if;
        if _type = 2 then
                insert into ybt_day_flow(unlock_minepool, date, create_time, update_time) values(obj.amount, _date, _now, _now) on conflict(date) do update
                set unlock_minepool = obj.amount, update_time=_now;                 
        elsif _type = 3 then
                insert into ybt_day_flow(unlock_project, date, create_time, update_time) values(obj.amount, _date, _now, _now) on conflict(date) do update
                set unlock_project = obj.amount, update_time=_now;   
        end if;
    end if;
         
    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建ybt解锁触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_ybt_flow on ybt_flow;
CREATE TRIGGER tg_to_update_ybt_flow
after INSERT OR UPDATE of amount
ON ybt_flow  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_ybt_flow_value();









-- 创建ybt解锁触发器
-- 创建ybt解锁触发器函数 
CREATE OR REPLACE FUNCTION fn_ybt_flow_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE                                                                                                                       
    obj record;                                                                                                               
    _type int;                                                                                                                
    _date varchar(20);                                                                                                        
    _now int;                                                                                                                 
BEGIN                                                                                                                         
    IF (TG_OP = 'DELETE') THEN                                                                                               
        _type = OLD.type;                                                                                                     
        _date = OLD.date;                                                                                                     
    ELSE                                                                                                                     
        _type = NEW.type;                                                                                                     
        _date = NEW.date;                                                                                                     
    END IF;                                                                                                                  
    _now = unix_timestamp();                                                                                                 
    -- ybt解锁                                                                                                                
    if _type = 0 or _type = 1 then    -- 空投及活动赠送奖励都算空投奖励                                                       
            select sum(amount) as amount into obj from ybt_flow where type in (0, 1) and date=_date;                          
            if obj is not null then                                                                                           
                insert into ybt_day_flow(reward, date, create_time, update_time) values(obj.amount, _date, _now, _now) on conflict(date) do update         
                set reward = obj.amount, update_time=_now;                                                                    
            end if;                                                                                                           
    else                                                                                                                      
        select sum(amount) as amount into obj from ybt_flow where type=_type and date=_date;            
        if obj is null THEN                                                                                                   
            return null;                                                                                                      
        end if;                                                                                                               
        if _type = 2 then                                                                                                     
                insert into ybt_day_flow(unlock_minepool, date, create_time, update_time) values(obj.amount, _date, _now, _now) on conflict(date) do update
                set unlock_minepool = obj.amount, update_time=_now;                                                           
        elsif _type = 3 then                                                                                                  
                insert into ybt_day_flow(unlock_project, date, create_time, update_time) values(obj.amount, _date, _now, _now) on conflict(date) do update 
                set unlock_project = obj.amount, update_time=_now;                                                            
        end if;        
    end if;                                                                                                                   
    RETURN null;                                                                                                              
END;      
$pt_to_update_value$ LANGUAGE plpgsql;


-- 创建ybt解锁触发器，注意要为后触发，行级触发器
drop trigger tg_to_delete_ybt_flow on ybt_flow;
CREATE TRIGGER tg_to_delete_ybt_flow
after DELETE
ON ybt_flow  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_ybt_flow_value();






-- 创建资金池状态触发器 并修改相应订单状态,添加分红记录,添加卖家打款操作
-- 创建金池状态触发器函数
CREATE OR REPLACE FUNCTION fn_assetpool_status_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  obj record;
  _date varchar(20);
  _now int;
  _poolid bigint;
  _status int;
BEGIN
    -- 支付状态信息
    _status = NEW.status;
    _poolid = NEW.id;
    _date = current_date;
    --_now = unix_timestamp();
    _now = NEW.update_time;   -- 用unix_timestamp()会有问题 可能与上次调用的值相同
    
    IF NEW.currency_type <> 1 or NEW.publish_area <> 1 THEN   -- 非KT专区
      return null;
    END IF;

    IF (TG_OP = 'UPDATE') THEN
        IF OLD.status != 0 THEN
            RAISE INFO 'fn_assetpool_status_value old status not 0';
            return null;
        END IF;
        IF NEW.status = 0 THEN
            RAISE INFO 'fn_assetpool_status_value update new status is 0';
            return null;            
        END IF;
    END IF;

    IF (TG_OP = 'INSERT') THEN
        IF NEW.status != 0 THEN
            RAISE INFO 'fn_assetpool_status_value insert status not 0';
            return null;
        END IF;
    END IF;

    -- 资金池的帐户不享受分红和收益
    IF _status = 0 THEN
      -- 划款给平台记录
      insert into user_asset_detail(user_id, type, transaction_type,amount,date, create_time,update_time) values(NEW.payer_userid, 1, 3, -NEW.pay_amount,_date,_now,_now);            
      --insert into user_asset_detail(user_id, type, transaction_type,amount,date, create_time,update_time) values(1, 1, 3, NEW.seller_amount,_date,now,now);            
    ELSIF _status = 1 THEN
      -- 划款给卖家记录
      --insert into user_asset_detail(user_id, type, transaction_type,amount,date, create_time,update_time) values(1, 1, 3, -NEW.seller_amount,_date,now,now);            
        insert into user_asset_detail(user_id, type, transaction_type,amount,date, create_time,update_time) values(NEW.seller_userid, 1, 4, NEW.seller_amount, _date,_now,_now);
    ELSIF _status = 2 THEN
      -- 划款给买家记录
      --insert into user_asset_detail(user_id, type, transaction_type,amount,date, create_time,update_time) values(1, 1, 3, -NEW.seller_amount,_date,now,now);            
      insert into user_asset_detail(user_id, type, transaction_type,amount,date, create_time,update_time) values(NEW.payer_userid, 1, 5, NEW.seller_amount, _date,_now,_now);
    END IF;          

    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建金池状态触发器，注意要为后触发，行级触发器
drop trigger tg_to_update_assetpool_status on yunbay_asset_pool;
CREATE TRIGGER tg_to_update_assetpool_status
after INSERT OR UPDATE of status    -- 后触发
--BEFORE UPDATE 
ON yunbay_asset_pool  -- 指定触发表
FOR EACH ROW     -- 行级触发
EXECUTE PROCEDURE fn_assetpool_status_value();



-- 创建平台资产触发器f
-- 创建平台资产触发器函数
CREATE OR REPLACE FUNCTION fn_yunbay_asset_value()
RETURNS TRIGGER
AS $pt_to_update_value$
DECLARE
  obj record;
  dt varchar(20);
  _now int;
BEGIN
    -- 自动更新平台资产表
    _now = unix_timestamp();
    dt = current_date;

    -- 查询当前平台所有资产信息并更新到平台资产表中
    select sum(amount) as amount, sum(profit) as profit, sum(issue_ybt) as issue_ybt,sum(destoryed_ybt) as destoryed_ybt, sum(air_recover) as air_recover, sum(perynbay) as perynbay into obj from yunbay_asset_detail;
    if obj is not null then
      insert into yunbay_asset(total_kt,total_kt_profit,total_issue_ybt,total_destroyed_ybt,total_air_recover,total_perynbay,date,create_time,update_time) 
      values(obj.amount, obj.profit, obj.issue_ybt, obj.destoryed_ybt, obj.air_recover, obj.perynbay, dt, _now, _now) on conflict(date) do update
      set total_kt=obj.amount, total_kt_profit=obj.profit, total_issue_ybt=obj.issue_ybt, total_destroyed_ybt=obj.destoryed_ybt, total_air_recover=obj.air_recover, total_perynbay=obj.perynbay;
    end if;
    RETURN null;
END;
$pt_to_update_value$ LANGUAGE plpgsql;

-- 创建平台资产明细触发器，注意要为后触发，行级触发器

drop trigger tg_to_update_yunbay_asset on yunbay_asset_detail;
CREATE TRIGGER tg_to_update_yunbay_asset
after INSERT OR DELETE OR UPDATE 
ON yunbay_asset_detail  -- 指定触发表
FOR EACH STATEMENT     -- 语句触发
EXECUTE PROCEDURE fn_yunbay_asset_value();