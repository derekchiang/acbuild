package acb

// This is the same default that runc uses
const LibcontainerDefaultConfig = `{
    "no_pivot_root": false,
    "parent_death_signal": 0,
    "pivot_dir": "",
    "rootfs": "/home/vagrant/playground/go/rootfs",
    "readonlyfs": true,
    "privatefs": true,
    "mounts": [
        {
            "source": "proc",
            "destination": "/proc",
            "device": "proc",
            "flags": 0,
            "data": "",
            "relabel": "",
            "premount_cmds": null,
            "postmount_cmds": null
        },
        {
            "source": "tmpfs",
            "destination": "/dev",
            "device": "tmpfs",
            "flags": 16777218,
            "data": "mode=755,size=65536k",
            "relabel": "",
            "premount_cmds": null,
            "postmount_cmds": null
        },
        {
            "source": "devpts",
            "destination": "/dev/pts",
            "device": "devpts",
            "flags": 10,
            "data": "newinstance,ptmxmode=0666,mode=0620,gid=5",
            "relabel": "",
            "premount_cmds": null,
            "postmount_cmds": null
        },
        {
            "source": "shm",
            "destination": "/dev/shm",
            "device": "tmpfs",
            "flags": 14,
            "data": "mode=1777,size=65536k",
            "relabel": "",
            "premount_cmds": null,
            "postmount_cmds": null
        },
        {
            "source": "mqueue",
            "destination": "/dev/mqueue",
            "device": "mqueue",
            "flags": 14,
            "data": "",
            "relabel": "",
            "premount_cmds": null,
            "postmount_cmds": null
        },
        {
            "source": "sysfs",
            "destination": "/sys",
            "device": "sysfs",
            "flags": 15,
            "data": "",
            "relabel": "",
            "premount_cmds": null,
            "postmount_cmds": null
        },
        {
            "source": "cgroup",
            "destination": "/sys/fs/cgroup",
            "device": "cgroup",
            "flags": 2097167,
            "data": "",
            "relabel": "",
            "premount_cmds": null,
            "postmount_cmds": null
        }
    ],
    "devices": [
        {
            "type": 99,
            "path": "/dev/null",
            "major": 1,
            "minor": 3,
            "permissions": "rwm",
            "file_mode": 8630,
            "uid": 0,
            "gid": 0
        },
        {
            "type": 99,
            "path": "/dev/random",
            "major": 1,
            "minor": 8,
            "permissions": "rwm",
            "file_mode": 8630,
            "uid": 0,
            "gid": 0
        },
        {
            "type": 99,
            "path": "/dev/full",
            "major": 1,
            "minor": 7,
            "permissions": "rwm",
            "file_mode": 8630,
            "uid": 0,
            "gid": 0
        },
        {
            "type": 99,
            "path": "/dev/tty",
            "major": 5,
            "minor": 0,
            "permissions": "rwm",
            "file_mode": 8630,
            "uid": 0,
            "gid": 5
        },
        {
            "type": 99,
            "path": "/dev/zero",
            "major": 1,
            "minor": 5,
            "permissions": "rwm",
            "file_mode": 8630,
            "uid": 0,
            "gid": 0
        },
        {
            "type": 99,
            "path": "/dev/urandom",
            "major": 1,
            "minor": 9,
            "permissions": "rwm",
            "file_mode": 8630,
            "uid": 0,
            "gid": 0
        }
    ],
    "mount_label": "",
    "hostname": "shell",
    "namespaces": [
        {
            "type": "NEWPID",
            "path": ""
        },
        {
            "type": "NEWNET",
            "path": ""
        },
        {
            "type": "NEWIPC",
            "path": ""
        },
        {
            "type": "NEWUTS",
            "path": ""
        },
        {
            "type": "NEWNS",
            "path": ""
        }
    ],
    "capabilities": [
        "AUDIT_WRITE",
        "KILL",
        "NET_BIND_SERVICE"
    ],
    "networks": null,
    "routes": null,
    "cgroups": {
        "name": "go",
        "parent": "/user.slice/user-1000.slice/session-2.scope",
        "allow_all_devices": false,
        "allowed_devices": [
            {
                "type": 99,
                "path": "/dev/null",
                "major": 1,
                "minor": 3,
                "permissions": "rwm",
                "file_mode": 8630,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 99,
                "path": "/dev/random",
                "major": 1,
                "minor": 8,
                "permissions": "rwm",
                "file_mode": 8630,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 99,
                "path": "/dev/full",
                "major": 1,
                "minor": 7,
                "permissions": "rwm",
                "file_mode": 8630,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 99,
                "path": "/dev/tty",
                "major": 5,
                "minor": 0,
                "permissions": "rwm",
                "file_mode": 8630,
                "uid": 0,
                "gid": 5
            },
            {
                "type": 99,
                "path": "/dev/zero",
                "major": 1,
                "minor": 5,
                "permissions": "rwm",
                "file_mode": 8630,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 99,
                "path": "/dev/urandom",
                "major": 1,
                "minor": 9,
                "permissions": "rwm",
                "file_mode": 8630,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 99,
                "path": "",
                "major": -1,
                "minor": -1,
                "permissions": "m",
                "file_mode": 0,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 98,
                "path": "",
                "major": -1,
                "minor": -1,
                "permissions": "m",
                "file_mode": 0,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 99,
                "path": "/dev/console",
                "major": 5,
                "minor": 1,
                "permissions": "rwm",
                "file_mode": 0,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 99,
                "path": "/dev/tty0",
                "major": 4,
                "minor": 0,
                "permissions": "rwm",
                "file_mode": 0,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 99,
                "path": "/dev/tty1",
                "major": 4,
                "minor": 1,
                "permissions": "rwm",
                "file_mode": 0,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 99,
                "path": "",
                "major": 136,
                "minor": -1,
                "permissions": "rwm",
                "file_mode": 0,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 99,
                "path": "",
                "major": 5,
                "minor": 2,
                "permissions": "rwm",
                "file_mode": 0,
                "uid": 0,
                "gid": 0
            },
            {
                "type": 99,
                "path": "",
                "major": 10,
                "minor": 200,
                "permissions": "rwm",
                "file_mode": 0,
                "uid": 0,
                "gid": 0
            }
        ],
        "denied_devices": null,
        "memory": 0,
        "memory_reservation": 0,
        "memory_swap": 0,
        "kernel_memory": 0,
        "cpu_shares": 0,
        "cpuset_cpus": "",
        "cpuset_mems": "",
        "blkio_throttle_read_bps_device": "",
        "blkio_throttle_write_bps_device": "",
        "blkio_throttle_read_iops_device": "",
        "blkio_throttle_write_iops_device": "",
        "blkio_weight": 0,
        "blkio_weight_device": "",
        "freezer": "",
        "hugetlb_limit": null,
        "slice": "",
        "oom_kill_disable": false,
        "memory_swappiness": 0,
        "net_prio_ifpriomap": null,
        "net_cls_classid": ""
    },
    "apparmor_profile": "",
    "process_label": "",
    "rlimits": null,
    "additional_groups": null,
    "uid_mappings": null,
    "gid_mappings": null,
    "mask_paths": [
        "/proc/kcore"
    ],
    "readonly_paths": [
        "/proc/sys",
        "/proc/sysrq-trigger",
        "/proc/irq",
        "/proc/bus"
    ],
    "sysctl": null,
    "seccomp": null
}`
