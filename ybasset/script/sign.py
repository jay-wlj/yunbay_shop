# coding=utf-8
import os
import urllib
import hashlib
import requests


def sign_request(uri, get_arg_dict, header_dict, post_body_str, sign_key):
    '''
    获取请求签名
    '''
    uri = urllib.quote(uri)
    args = [(k, get_arg_dict[k]) for k in sorted(get_arg_dict.keys())]
    args_str = urllib.urlencode(args).replace('+', '%20')
    h_keys_set = set([k.upper().startswith('X-YF') and k or None for k in header_dict.keys()])
    if None in h_keys_set:
        h_keys_set.remove(None)
    h_keys = list(h_keys_set)
    h_keys.sort(key=lambda x: x.lower())
    headers_str = '\n'.join(['%s:%s' % (k.lower(), (header_dict[k] or "").strip()) for k in h_keys])
    header_keys_str = ';'.join([k.lower() for k in h_keys])
    sha1_body = hashlib.sha1(post_body_str or "").hexdigest()
    sign_str = '\n'.join([uri, args_str, headers_str, header_keys_str, sha1_body, sign_key])
    # print sign_str
    return hashlib.sha1(sign_str).hexdigest()