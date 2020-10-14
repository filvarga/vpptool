#/usr/bin/env python3

from scapy.all import sendp
from scapy.layers.inet import IP, TCP
from scapy.layers.l2 import Ether, Dot1Q

import netifaces as ni
from random import randrange

def get_ip(iface):
    return ni.ifaddresses(iface)[ni.AF_INET][0]['addr']

def build_tcp_packet(iface, dst_ip, dst_port):
    return (Ether(dst="00:00:00:03:00:00") /
            IP(src=get_ip(iface), dst=dst_ip, ttl=64) /
            TCP(sport=randrange(0, 65535), dport=dst_port))

def send_tcp_packet(iface, dst_ip, dst_port):
    p = build_tcp_packet(iface, dst_ip, dst_port)
    p.show2()
    sendp(p, iface=iface)

iface="host0"

send_tcp_packet(iface, "20.0.1.1", 5002)
