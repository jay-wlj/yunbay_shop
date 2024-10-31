/**
CREATE USER ybapi WITH PASSWORD '123456';
CREATE DATABASE ybapi with owner=ybapi ENCODING='UTF8';
GRANT ALL PRIVILEGES ON DATABASE ybapi to ybapi;
\c ybapi ybapi;
**/


create table uploadfile(
  id bigserial,
  rid varchar(64) not null,
  appid varchar(32) not null,
  "hash" varchar(64) not null,
  size int default 0,
  path varchar(256) not null,
  width int default 0,
  height int default 0,
  duration int default 0,
  extinfo jsonb null,
  create_time int not null,
  update_time int not null,
  primary key(id)
);



COMMENT ON TABLE uploadfile IS '文件信息表';
COMMENT ON COLUMN uploadfile.rid IS '资源ID';
COMMENT ON COLUMN uploadfile.appid IS '资源所属app';
COMMENT ON COLUMN uploadfile.hash IS '资源HASH';
COMMENT ON COLUMN uploadfile.size IS '资源大小';
COMMENT ON COLUMN uploadfile.path IS '资源URL';
COMMENT ON COLUMN uploadfile.width IS '图片视频宽度(像素)';
COMMENT ON COLUMN uploadfile.height IS '图片视频高度(像素)';
COMMENT ON COLUMN uploadfile.duration IS '视频时长';
COMMENT ON COLUMN uploadfile.size IS '资源大小';
COMMENT ON COLUMN uploadfile.extinfo IS '扩展信息';

CREATE UNIQUE INDEX uploadfile_rid ON uploadfile(rid);
CREATE INDEX uploadfile_create_time_idx ON uploadfile(create_time);
CREATE INDEX uploadfile_hash ON uploadfile("hash");
CREATE INDEX uploadfile_appid ON uploadfile(appid);

