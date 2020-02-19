#!/usr/bin/python
# -*- coding: UTF-8 -*-

import json

def helloWord():
    data = {
        "msgtype": "markdown",
        "markdown": {
            "content": '### 这是一个标题\n\n > 你好世界'
        }
    }
    print(json.dumps(data))

if __name__ == "__main__":
    helloWord()
