#!/usr/bin/env python3

import json
import urllib.request
import sys

GOLD_PRICE_URL = "https://www.freeforexapi.com/api/live?pairs=USDXAU"


def make_json_request(url):
    req = urllib.request.Request(url)
    req.add_header(
        "User-Agent",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36",
    )
    return urllib.request.urlopen(req).read()


def main():
    random = make_json_request(GOLD_PRICE_URL)
    return 1 / float(json.loads(random)["rates"]["USDXAU"]["rate"])


if __name__ == "__main__":
    try:
        print(main(*sys.argv[1:]))
    except Exception as e:
        print(str(e), file=sys.stderr)
        sys.exit(1)
