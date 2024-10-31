\c ybaccount ybaccount


\c ybapi ybapi



\c ybasset ybasset




\c ybapi ybapi

-- 修改 运营管理
update admin_controller set name='运营管理(国际)' where id=1;
-- 修改 运营管理
update admin_controller set name='运营管理(国际)' where id=1;

-- 新增权限
insert into admin_controller (id, module, controller, name, create_time, update_time) values(8, 'backend', 'OperationCn', '运营管理(国内)', '1542104757', 1542104757);
insert into "admin_action" values ('89', 8, 'SaveRecommend', '/backend/operation-cn/save-recommend', '平台优选/最新发布  编辑', 1542104757, 1542104757);
insert into "admin_action" values (90, 8, 'Recommend', '/backend/operation-cn/recommend', '平台优选/最新发布  列表', 1542104757, 1542104757);
insert into "admin_action" values (91, 8, 'GetBanner', '/backend/operation-cn/get-banner', 'banner运营 列表', 1542104757, 1542104757);
insert into "admin_action" values (92, 8, 'SaveBanner', '/backend/operation-cn/save-banner', 'banner运营 编辑', 1542104757, 1542104757);
insert into "admin_action" values (93, 8, 'ListNotice', '/backend/operation-cn/list-notice', '资讯/公告 列表', 1542104757, 1542104757);
insert into "admin_action" values (94, 8, 'EditNotive', '/backend/operation-cn/edit-notice', '资讯/公告 编辑', 1542104757, 1542104757);
insert into "admin_action" values (95, 8, 'AddNotice', '/backend/operation-cn/add-notice', '资讯/公告 发布', 1542104757, 1542104757);
insert into "admin_action" values (96, 8, 'DelNotice', '/backend/operation-cn/del-notice', '资讯/公告 删除', 1542104757, 1542104757);
insert into "admin_action" values (97, 8, 'RecommendNotice', '/backend/operation-cn/recommend-notice', '资讯/公告 推荐/取消推荐', 1542104757, 1542104757);
insert into "admin_action" values (98, 8, 'GetUpgrade', '/backend/operation-cn/get-upgrade', '查看升级设置', 1542104757, 1542104757);
insert into "admin_action" values (99, 8, 'SaveUpgrade', '/backend/operation-cn/save-upgrade', '保存升级设置', 1542104757, 1542104757);
insert into "admin_action" values (100, 6, 'DrawSet', '/backend/wallet/draw-set', '国内提现处理', 1542104757, 1542104757);
insert into "admin_action" values (101, 4, 'SaveSnetPrice', '/backend/setting/save-snet-price', 'SNET单价值设置', 1542104757, 1542104757);

insert into "admin_action" values (102, 1, 'DiscountList', '/backend/operation/discount-list', '折扣专区  列表', 1545735932, 1545735932);
insert into "admin_action" values (103, 1, 'SaveDiscount', '/backend/operation/save-discount', '折扣专区  编辑', 1545735932, 1545735932);
insert into "admin_controller" (id, module, controller, name, create_time, update_time) values (9, 'backend', 'Report', '报表',  1546935284, 1546935284);
insert into "admin_action" values (104, 9, 'YbassetList', '/backend/report/ybasset-list', '每日数据报表 列表', 1546935284, 1546935284);
insert into "admin_action" values (105, 1, 'EditVoucher', '/backend/product/edit-voucher', '编辑代金券', 1546935284, 1546935284);
insert into "admin_action" values (106, 1, 'ListVoucher', '/backend/product/list-voucher', '代金券列表', 1546935284, 1546935284);  



INSERT INTO "public"."admin_action" VALUES ('86', '4', 'Check', '/backend/product/check', 'YBT专区 商品审核', '1542104757', '1542104757');
INSERT INTO "public"."admin_action" VALUES ('87', '4', 'SaveYbtPrice', '/backend/setting/save-ybt-price', 'YBT单价值设置', '1542104757', '1542104757');
INSERT INTO "public"."admin_action" VALUES ('88', '6', 'SystemDraw', '/backend/wallet/system-draw', '账户转账', '1542104757', '1542104757');




-- 新增权限
insert into admin_controller (id, module, controller, name, create_time, update_time) values(8, 'backend', 'OperationCn', '运营管理(国内)', '1542104757', 1542104757);
insert into "admin_action" values ('89', 8, 'SaveRecommend', '/backend/operation-cn/save-recommend', '平台优选/最新发布  编辑', 1542104757, 1542104757);
insert into "admin_action" values (90, 8, 'Recommend', '/backend/operation-cn/recommend', '平台优选/最新发布  列表', 1542104757, 1542104757);
insert into "admin_action" values (91, 8, 'GetBanner', '/backend/operation-cn/get-banner', 'banner运营 列表', 1542104757, 1542104757);
insert into "admin_action" values (92, 8, 'SaveBanner', '/backend/operation-cn/save-banner', 'banner运营 编辑', 1542104757, 1542104757);
insert into "admin_action" values (93, 8, 'ListNotice', '/backend/operation-cn/list-notice', '资讯/公告 列表', 1542104757, 1542104757);
insert into "admin_action" values (94, 8, 'EditNotive', '/backend/operation-cn/edit-notice', '资讯/公告 编辑', 1542104757, 1542104757);
insert into "admin_action" values (95, 8, 'AddNotice', '/backend/operation-cn/add-notice', '资讯/公告 发布', 1542104757, 1542104757);
insert into "admin_action" values (96, 8, 'DelNotice', '/backend/operation-cn/del-notice', '资讯/公告 删除', 1542104757, 1542104757);
insert into "admin_action" values (97, 8, 'RecommendNotice', '/backend/operation-cn/recommend-notice', '资讯/公告 推荐/取消推荐', 1542104757, 1542104757);
insert into "admin_action" values (98, 8, 'GetUpgrade', '/backend/operation-cn/get-upgrade', '查看升级设置', 1542104757, 1542104757);
insert into "admin_action" values (99, 8, 'SaveUpgrade', '/backend/operation-cn/save-upgrade', '保存升级设置', 1542104757, 1542104757);
insert into "admin_action" values (100, 6, 'DrawSet', '/backend/wallet/draw-set', '国内提现处理', 1542104757, 1542104757);
insert into "admin_action" values (101, 4, 'SaveSnetPrice', '/backend/setting/save-snet-price', 'SNET单价值设置', 1542104757, 1542104757);
INSERT INTO "public"."admin_action" VALUES ('86', '4', 'Check', '/backend/product/check', 'YBT专区 商品审核', '1542104757', '1542104757');
INSERT INTO "public"."admin_action" VALUES ('87', '4', 'SaveYbtPrice', '/backend/setting/save-ybt-price', 'YBT单价值设置', '1542104757', '1542104757');
INSERT INTO "public"."admin_action" VALUES ('88', '6', 'SystemDraw', '/backend/wallet/system-draw', '账户转账', '1542104757', '1542104757');




-- 新增公共账号
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (17, 'YBT专区销售进账', '11', '0x7fe94ae4b04693f43c9087386a5211d007a29233', 0, 1542104757, 1542104757);
insert into common_account_address (id, name, uid, address, type, create_time, update_time) values (18, 'YBT专区KT支出', '12', '0xc6d3ab85af87ca8c9d3ce9c192135ba0ab599145', 0, 1542104757, 1542104757);





