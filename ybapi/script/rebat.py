# coding=utf-8

import psycopg2
import re
import sys
import time
import math
import json
import requests
import argparse
import ConfigParser
import os


# 添加用户ybt分红明细
def rebat_user_asset_detail(cursor, user_id, ybt_amount, date):
    sql = "insert into user_asset_detail(user_id, type, transaction_type, amount,date,create_time,update_time) values(%d, %d, %d, %f,"
    cursor.execute(sql)
    rows = cursor.fetchall()
    map_types = {}
    if rows:
        for row in rows:
            type_str = str(row[0])
            if not map_types.has_key(type_str):
                map_types[type_str] = []
            map_types[type_str].append(row[1])
  
    for k in map_types:
        ids = map_types[k]
        map_user_ids = {}
        for res_id in ids:
            user_id = get_res_user_id(apicursor, int(k), res_id)
            if user_id >0 :
                map_user_ids[str(res_id)] = user_id

        for resid in map_user_ids:
            save_res_user_id(cursor, int(k), int(resid), map_user_ids[resid])

def rebat_asset(cursor):
    # 获取昨日平台交易池数据


if __name__ == '__main__':
    begin_time = time.time()

    parser = argparse.ArgumentParser(description="平台ybt分红脚本")

    parser.add_argument('-host', dest='psql_host', default="127.0.0.1", help="psql ip地址")
    parser.add_argument('-p', dest="psql_port", default=5432, type=int, help="psql端口号")
    parser.add_argument('-u', dest="psql_user", default="nfapi", help="psql用户名")
    parser.add_argument('-passwd', dest="psql_passwd", default="123456", help="psql密码")
    parser.add_argument('-db', dest="psql_db", default="nfapi", help="psql db名字")
    args = parser.parse_args()

    api_db = psycopg2.connect(host=args.psql_host, user=args.psql_user, password=args.psql_passwd, database=args.psql_db, port=args.psql_port)
    cursor = api_db.cursor()


    syn_comment_res_userid(comment_cursor, cursor)

    api_db.commit()
    api_db.close()

