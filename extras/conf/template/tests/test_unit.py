#!/usr/bin/env python

import unittest
from socket import AF_INET, inet_pton
from scapy.layers.inet import IP, TCP
from scapy.layers.l2 import Ether, Dot1Q

from framework import VppTestCase, VppTestRunner
from vpp_sub_interface import VppDot1QSubint
from vpp_papi import VppEnum


class NAT44MethodHolder(VppTestCase):
    """ NAT create capture and verify method holder """

    @classmethod
    def nat_add_address(cls, ip, vrf_id=0xFFFFFFFF):
        cls.vapi.nat44_add_del_address_range(first_ip_address=ip,
                                              last_ip_address=ip,
                                              vrf_id=vrf_id,
                                              is_add=1,
                                              flags=0)

    @classmethod
    def nat_add_inside_interface(cls, i):
        flags = VppEnum.vl_api_nat_config_flags_t.NAT_IS_INSIDE
        print(cls.vapi.cli("show int"))
        print(i.sw_if_index)
        cls.vapi.nat44_interface_add_del_feature(
            sw_if_index=i.sw_if_index,
            flags=flags, is_add=1)

    @classmethod
    def nat_add_outside_interface(cls, i):
        flags = VppEnum.vl_api_nat_config_flags_t.NAT_IS_OUTSIDE
        cls.vapi.nat44_interface_add_del_feature(
            sw_if_index=i.sw_if_index,
            flags=flags, is_add=1)


class TestSetup(NAT44MethodHolder):
    """ Setup test cases """

    @classmethod
    def setUpClass(cls):
        super(TestSetup, cls).setUpClass()

        try:
            cls.vapi.ip_table_add_del(is_add=1, table={'table_id': 10})
            cls.vapi.ip_table_add_del(is_add=1, table={'table_id': 20})
            cls.vapi.ip_table_add_del(is_add=1, table={'table_id': 30})

            cls.create_pg_interfaces(range(2))

            cls.pg0.admin_up()
            cls.pg1.admin_up()

            # ip4-not-enabled (required for subinterfaces ?)
            cls.pg0.config_ip4()
            cls.pg1.config_ip4()

            cls.vlan1001 = VppDot1QSubint(cls, cls.pg0, 1001)
            cls.vlan1001.set_table_ip4(10)
            cls.vlan1001.admin_up()

            cls.vlan1002 = VppDot1QSubint(cls, cls.pg0, 1002)
            cls.vlan1002.set_table_ip4(20)
            cls.vlan1002.admin_up()

            cls.vlan1003 = VppDot1QSubint(cls, cls.pg1, 1003)
            cls.vlan1003.set_table_ip4(30)
            cls.vlan1003.admin_up()

            cls.vlan1001.config_ip4()
            cls.vlan1001.resolve_arp()
            cls.vlan1001.generate_remote_hosts(1)
            cls.vlan1001.configure_ipv4_neighbors()

            cls.vlan1002.config_ip4()
            cls.vlan1002.resolve_arp()
            cls.vlan1002.generate_remote_hosts(1)
            cls.vlan1002.configure_ipv4_neighbors()

            cls.vlan1003.config_ip4()
            cls.vlan1003.resolve_arp()
            cls.vlan1003.generate_remote_hosts(3)
            cls.vlan1003.configure_ipv4_neighbors()

            cls.nat_add_inside_interface(cls.vlan1001)
            cls.nat_add_inside_interface(cls.vlan1002)
            cls.nat_add_outside_interface(cls.vlan1003)

            cls.nat_add_address(cls.vlan1003._remote_hosts[0]._ip4, vrf_id=10)
            cls.nat_add_address(cls.vlan1003._remote_hosts[1]._ip4, vrf_id=20)

            print(cls.vapi.cli("show interface address"))
            print(cls.vapi.cli("show nat44 address"))

        except Exception:
            super(TestSetup, cls).tearDownClass()
            raise

    @classmethod
    def tearDownClass(cls):
        super(TestSetup, cls).tearDownClass()

    def test_setup(self):
        """ Setup test case """
        # where should the stream pass ??
        p1 = self.create_tcp_pkt(
            src_mac=self.vlan1001.remote_mac, dst_mac=self.vlan1001.local_mac,
            src_ip=self.vlan1001._remote_hosts[0]._ip4, src_port=40000,
            dst_ip=self.vlan1003._remote_hosts[2]._ip4, dst_port=80)
        p2 = self.create_tcp_pkt(
            src_mac=self.vlan1002.remote_mac, dst_mac=self.vlan1002.local_mac,
            src_ip=self.vlan1002._remote_hosts[0]._ip4, src_port=40000,
            dst_ip=self.vlan1003._remote_hosts[2]._ip4, dst_port=80)

        self.pg0.add_stream([p1,p2])
        self.pg_enable_capture(self.pg_interfaces)
        self.pg_start()

        print(self.vapi.cli("show errors"))
        capture = self.pg1.get_capture(2)
        print(capture)

    def tearDown(self):
        super(TestSetup, self).tearDown()

    @staticmethod
    def create_tcp_pkt(src_mac, dst_mac, src_port, dst_port, src_ip, dst_ip, ttl=64, vlan=0):
        return (Ether(src=src_mac, dst=dst_mac) /
                Dot1Q(vlan=vlan) /
                IP(src=src_ip, dst=dst_ip, ttl=ttl) /
                TCP(sport=src_port, dport=dst_port))

    def Xtest_tcp_handshake(self, in_if, out_if):
        # Test TCP 3 way handshake
        try:
            # SYN packet in->out
            p = (Ether(src=in_if.remote_mac, dst=in_if.local_mac) /
                 IP(src=in_if.remote_ip4, dst=out_if.remote_ip4) /
                 TCP(sport=self.tcp_port_in, dport=self.tcp_external_port,
                     flags="S"))
            in_if.add_stream(p)
            self.pg_enable_capture(self.pg_interfaces)
            self.pg_start()
            capture = out_if.get_capture(1)
            p = capture[0]
            self.tcp_port_out = p[TCP].sport

            # SYN + ACK packet out->in
            p = (Ether(src=out_if.remote_mac, dst=out_if.local_mac) /
                 IP(src=out_if.remote_ip4, dst=self.nat_addr) /
                 TCP(sport=self.tcp_external_port, dport=self.tcp_port_out,
                     flags="SA"))
            out_if.add_stream(p)
            self.pg_enable_capture(self.pg_interfaces)
            self.pg_start()
            in_if.get_capture(1)

            # ACK packet in->out
            p = (Ether(src=in_if.remote_mac, dst=in_if.local_mac) /
                 IP(src=in_if.remote_ip4, dst=out_if.remote_ip4) /
                 TCP(sport=self.tcp_port_in, dport=self.tcp_external_port,
                     flags="A"))
            in_if.add_stream(p)
            self.pg_enable_capture(self.pg_interfaces)
            self.pg_start()
            out_if.get_capture(1)

        except:
            self.logger.error("TCP 3 way handshake failed")
            raise

if __name__ == '__main__':
    unittest.main(testRunner=VppTestRunner)
