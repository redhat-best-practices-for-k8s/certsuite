console.log("hello")
window.classification = {
    "classification": {
        "access-control-pod-role-bindings":[
            {
             "ForTelco": "Mandatory",
             "FarEdge" : "Mandatory",
             "ForNonTelco": "Optional",
             "ForVZ": "Mandatory"
            }
        ],
        "access-control-container-host-port":[
            {
             "ForTelco": "Mandatory",
             "FarEdge" : "Mandatory",
             "ForNonTelco": "Mandatory",
             "ForVZ": "Mandatory"
            }
        ],   
        "access-control-ipc-lock-capability-check":[
            {
             "ForTelco": "Mandatory",
             "FarEdge" : "Mandatory",
             "ForNonTelco": "Optional",
             "ForVZ": "Mandatory"
            }
        ],
        "access-control-namespace":[
            {
             "ForTelco": "Mandatory",
             "FarEdge" : "Mandatory",
             "ForNonTelco": "Mandatory",
             "ForVZ": "Mandatory"
            }
        ],    
        "access-control-namespace-resource-quota":[
            {
             "ForTelco": "Optional",
             "FarEdge" : "Optional",
             "ForNonTelco": "Optional",
             "ForVZ": "Mandatory"
            }
        ],
        "access-control-net-admin-capability-check":[
            {
             "ForTelco": "Mandatory",
             "FarEdge" : "Mandatory",
             "ForNonTelco": "Optional",
             "ForVZ": "Mandatory"
            }
        ],
        "access-control-net-raw-capability-check":[
            {
             "ForTelco": "Mandatory",
             "FarEdge" : "Mandatory",
             "ForNonTelco": "Optional",
             "ForVZ": "Mandatory"
            }
        ],
        "access-control-no-1337-uid":[
            {
             "ForTelco": "Optional",
             "FarEdge" : "Optional",
             "ForNonTelco": "Optional",
             "ForVZ": "Mandatory"
            }
        ],
        "access-control-one-process-per-container":[
            {
                "ForTelco": "Optional",
                "FarEdge" : "Optional",
                "ForNonTelco": "Optional",
                "ForVZ": "Optional"
               }
        ],
        "access-control-pod-automount-service-account-token":[
            {
                "ForTelco": "Mandatory",
                "FarEdge" : "Mandatory",
                "ForNonTelco": "Optional",
                "ForVZ": "Mandatory"
               }
        ],
        "access-control-pod-host-ipc":[
            {
                "ForTelco": "Mandatory",
                "FarEdge" : "Mandatory",
                "ForNonTelco": "Mandatory",
                "ForVZ": "Mandatory"
               }
        ],
        "access-control-pod-host-network":[
            {
                "ForTelco": "Mandatory",
                "FarEdge" : "Mandatory",
                "ForNonTelco": "Mandatory",
                "ForVZ": "Mandatory"
               }
        ],
        "access-control-pod-host-path": [
            {
                "ForTelco": "Mandatory",
                "FarEdge" : "Mandatory",
                "ForNonTelco": "Mandatory",
                "ForVZ": "Mandatory"
               }
        ],
        "access-control-pod-host-pid":[
            {
                "ForTelco": "Mandatory",
                "FarEdge" : "Mandatory",
                "ForNonTelco": "Mandatory",
                "ForVZ": "Mandatory"
               }
        ],
        "access-control-pod-role-bindings":[
            {
                "ForTelco": "Mandatory",
                "FarEdge" : "Mandatory",
                "ForNonTelco": "Mandatory",
                "ForVZ": "Mandatory"
               }
        ],
        "access-control-pod-service-account":[
            {
                "ForTelco": "Mandatory",
                "FarEdge" : "Mandatory",
                "ForNonTelco": "Mandatory",
                "ForVZ": "Mandatory"
               }
        ],
        "access-control-requests-and-limits":[
            {
                "ForTelco": "Mandatory",
                "FarEdge" : "Mandatory",
                "ForNonTelco": "Optional",
                "ForVZ": "Mandatory"
               }
        ],
        "access-control-security-context": [
            {
                "ForNonTelco": "Optional",
                "ForTelco": "Optional",
                "FarEdge": "Optional",
                "ForVZ": "Mandatory"
            }
        ],
        "access-control-security-context-non-root-user-check": [
            {
                "ForNonTelco": "Mandatory",
                "ForTelco": "Mandatory",
                "FarEdge": "Mandatory",
                "ForVZ": "Mandatory"
            }
        ],
        "access-control-security-context-privilege-escalation": [
            {
            "ForNonTelco": "Mandatory",
            "ForTelco": "Mandatory",
            "FarEdge": "Mandatory",
            "ForVZ": "Mandatory"
        }
        ],
        "access-control-ssh-daemons": [
            {
                "ForNonTelco": "Optional",
                "ForTelco": "Mandatory",
                "FarEdge": "Mandatory",
                "ForVZ": "Mandatory"
            }
        ],
        "access-control-sys-admin-capability-check": [
            {
                "ForNonTelco": "Mandatory",
                "ForTelco": "Mandatory",
                "FarEdge": "Mandatory",
                "ForVZ": "Mandatory"
            }
        ],
        "access-control-sys-nice-realtime-capability": [
            {
                "ForNonTelco": "Optional",
                "ForTelco": "Mandatory",
                "FarEdge": "Mandatory",
                "ForVZ": "Mandatory"
            }
        ],
        "access-control-sys-ptrace-capability": [
            {
            "ForNonTelco": "Optional",
            "ForTelco": "Mandatory",
            "FarEdge": "Mandatory",
            "ForVZ": "Mandatory"
        }
        ],
        "affiliated-certification-container-is-certified":[
            {
                "ForNonTelco": "Mandatory",
                "ForTelco": "Mandatory",
                "FarEdge": "Mandatory",
                "ForVZ": "Mandatory"
        }
    ],
    "affiliated-certification-container-is-certified-digest": [
        {
            "ForNonTelco": "Mandatory",
            "ForTelco": "Mandatory",
            "FarEdge": "Mandatory",
            "ForVZ": "Mandatory"
        }
    ]
        
        }
    }