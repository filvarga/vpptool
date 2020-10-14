from trex.astf.api import *


class Prof1():
    cps = 1
    def __init__(self):
        pass

    def get_profile(self, **kwargs):
        ip_gen_c = ASTFIPGenDist(ip_range=["20.0.0.3", "20.0.0.254"],
                distribution="seq")
        ip_gen_s = ASTFIPGenDist(ip_range=["20.0.1.3", "20.0.1.254"],
                distribution="seq")
        ip_gen = ASTFIPGen(glob=ASTFIPGenGlobal(ip_offset="1.0.0.0"),
                           dist_client=ip_gen_c,
                           dist_server=ip_gen_s)
        return ASTFProfile(default_ip_gen=ip_gen, 
                cap_list=[ASTFCapInfo(file="../avl/trex_test.pcap", cps=self.cps)])

def register():
    return Prof1()

