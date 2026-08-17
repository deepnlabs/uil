#!/usr/bin/env python3
"""
ROS2 UIL-X Safety & Emergency Stop Bridge Node
Listens to local uild socket interlock breaches and publishes ROS2 E-Stop signals.
"""

import json
import os
import socket
import sys

# Try importing rclpy; if unavailable, run in Standalone IPC Logging Mode
HAS_ROS2 = True
try:
    import rclpy
    from rclpy.node import Node
    from std_msgs.msg import Bool
    from geometry_msgs.msg import Twist
except ImportError:
    HAS_ROS2 = False


SOCKET_PATH = "/tmp/uild.sock"


def run_standalone_listener():
    """Fallback IPC listener for testing socket events when ROS2 environment is not sourced."""
    print("⚠️  [WARN] 'rclpy' not detected. Running in Standalone IPC Socket Listener Mode...")
    
    if not os.path.exists(SOCKET_PATH):
        print(f"❌ [ERROR] Socket path {SOCKET_PATH} does not exist. Ensure `uild` is running!")
        return

    sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    try:
        sock.connect(SOCKET_PATH)
        print(f"✅ Connected to uild socket at {SOCKET_PATH}. Listening for safety interlock events...\n")
        
        buffer = ""
        while True:
            data = sock.recv(1024).decode('utf-8')
            if not data:
                break
            buffer += data
            while '\n' in buffer:
                line, buffer = buffer.split('\n', 1)
                if line.strip():
                    payload = json.loads(line)
                    print(f"  └─ [EVENT RECEIVED] Interlock: {payload.get('interlock_id')} | Breach: {payload.get('breach')}")
    except Exception as e:
        print(f"❌ Socket error: {e}")
    finally:
        sock.close()


if HAS_ROS2:
    class UILEmergencyStopNode(Node):
        def __init__(self):
            super().__init__('uil_emergency_stop_bridge')
            self.estop_pub = self.create_publisher(Bool, '/safety/emergency_stop', 10)
            self.cmd_vel_pub = self.create_publisher(Twist, '/cmd_vel', 10)
            self.get_logger().info('UIL-X ROS2 Emergency Stop Bridge initialized.')

        def run_socket_listener(self):
            if not os.path.exists(SOCKET_PATH):
                self.get_logger().error(f'Socket file {SOCKET_PATH} does not exist.')
                return

            sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
            try:
                sock.connect(SOCKET_PATH)
                self.get_logger().info('Connected to uild IPC socket.')
                buffer = ""
                while rclpy.ok():
                    data = sock.recv(1024).decode('utf-8')
                    if not data:
                        break
                    buffer += data
                    while '\n' in buffer:
                        line, buffer = buffer.split('\n', 1)
                        if line.strip():
                            self.process_safety_event(json.loads(line))
            except Exception as e:
                self.get_logger().error(f'Socket error: {e}')
            finally:
                sock.close()

        def process_safety_event(self, event):
            if event.get('breach', False):
                self.get_logger().fatal(f"🚨 SAFETY BREACH: [{event.get('interlock_id')}]! Triggering ROS2 E-Stop.")
                msg = Bool()
                msg.data = True
                self.estop_pub.publish(msg)

                stop_cmd = Twist()
                self.cmd_vel_pub.publish(stop_cmd)


def main(args=None):
    if not HAS_ROS2:
        run_standalone_listener()
        return

    rclpy.init(args=args)
    node = UILEmergencyStopNode()
    try:
        node.run_socket_listener()
    except KeyboardInterrupt:
        pass
    finally:
        node.destroy_node()
        rclpy.shutdown()


if __name__ == '__main__':
    main()
