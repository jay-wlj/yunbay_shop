#coding=utf-8

import json
import time
from datetime import datetime
import traceback
import sys
import argparse
import psycopg2

def process_dislike(cursor, ncursor):
    sql = 'select user_id from reward_ybt where reason like ''注册奖励'''
    
    user_ids = []
    cursor.execute(sql)
    results = cursor.fetchall()
    for row in results:
        sql = 'insert into user_asset_detail(user_id, type, transactio_type, amount, date, create_time, update_time) values(%d,%d,%d,%f,%s,%d,%d)'%(row[0], 1, 1, 10000, '2018-09-10', 1536570329, 1536570329)
        cursor.execute(sql)        

def main():
    begin_time = time.time()

    parser = argparse.ArgumentParser(description="迁移旧版不喜欢数据到新版数据库")
    parser.add_argument('-host', dest='psql_host', default="127.0.0.1", help="psql ip地址")
    parser.add_argument('-p', dest="psql_port", default=5432, type=int, help="psql端口号")
    parser.add_argument('-u', dest="psql_user", default="root", help="psql用户名")
    parser.add_argument('-passwd', dest="psql_passwd", default="123456", help="psql密码")
    parser.add_argument('-db', dest="psql_db", default="ybasset", help="psql db名字")

    args = parser.parse_args()

    db = psycopg2.connect(host=args.psql_host, user=args.psql_user, password=args.psql_passwd, database=args.psql_db, port=args.psql_port)

    cursor = db.cursor()

    process_dislike(cursor, ncursor)
    db.commit()

    

if __name__ == '__main__':
    main()