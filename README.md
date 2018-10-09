# GOAR (Go Automatic remediation):

GOAR is designed to handle event-based workflows on the networks, is a sample system built on basic scale principles constructed in Go - a static typing language with concurrency built in its philosophy. It uses RabbitMQ, as a communication channel between microservice-like instances.

The system allows setting up different event sources (data trailers) that are then processed (processors) and converted into events that translate into the execution of a job (executor). 

Events are the most critical part of the system, this are created by a processor and is handled by the executor, they translate into a combination of pre/post audits or simple jobs execution. This fundamental mechanism allows for the creation of the most common execution pipelines in the network.

The jobs/audits are independent piece of code built in any language (By default in the remediation/ folder), as long as they handle correctly receiving the arguments defined in the rules and output a pre-define JSON struct (define in ProcessOutput) anything should be callable (Perl, Python, Golang, etc.).

# Quick walkthrough 

Critical Syslog is triggered, a file tailer picks it up and send it for processing - RabbitMQ (together with all the message that the device is generating), a processor reads each message and catches the interesting critical message, it creates an event with the configured rule and what jobs are to be executed, the executor picks up this event and runs what is suppose to do. All the messages are handled by RabbitMQ, in JSON format (simple troubleshooting)

## Examples: 

- Parity error message is created by RouterA via syslog, its fast-handled by a tailer and read from the Syslog processor, it then creates an event with two jobs (defined in the rules): A capacity audit, and then if that's ok, a config push to RouterA that sets the metric of OSPF to maximum effectively draining the device from traffic.

- Interface goes down, Syslog tailer picks up the message and pushes to the queue, a processor create an event using the rule specification, with an audit that looks for common issues on an interface, a capacity audit and a config push to modify the metric of the link. The first audit will determine if we have issues on the interface like flaps, error, etc., the capacity audit will make sure we have enough capacity to move around the traffic and not leave anything disconnected, and the config push will remove production traffic from the interface.

<pre>
                                                                                                                         
                                           - RuleName: arista_interface_down                                             
                                             DeviceType: ARISTA                                                          
                                             Regex: 'Line protocol on Interface                                          
                                           (?P<interface>\S+).+changed state to                                          
                                           down'                                                                         
                                             PreAudits:                                                                  
                                               - interface_check.py                                                      
                                               - capacity_link.py                                                        
                                             Remediations:                                                               
                                               - port_down_arista.py                                                     
 test_device Ebra: 1417:                     AlertType: Interface Status                                                 
 %LINEPROTO-5-UPDOWN: Line protocol on                                                                                   
 Interface Ethernet6/12/1, changed state                                                                                 
 to down                                                       │                                                         
                                                               │                          ┏━━━━━━━━━━━━━━━━┓             
      ┌────────┐              ┏━━━━━━━━━━━━┓          ┌────────▼──────────┐               ┃                ┃             
      │        │              ┃            ┃          │                   │               ┃                ┃             
      │ Syslog │─────────────▶┃  RabbitMQ  ◀──────────│ Syslog Processor  ├──────────────▶┃    RabbitMQ    ┃             
      │        │              ┃ QUEUE_LOG  ┃          │                   │               ┃ QUEUE_INCIDENT ┃             
      └────────┘              ┃            ┃          └───────────────────┘               ┃                ┃             
                              ┗━━━━━━━━━━━━┛                                              ┃                ┃             
                                                                                          ┗━━━━━━━▲━━━━━━━━┛             
                                                                                                  │                      
                                                                                                  │                      
                                                                                                  │                      
                                                                                           ┌──────┴───────┐              
                                                                                           │              │              
                                                                                           │   EXECUTOR   │              
                                                                                           │              │              
                                                                                           └──┬────┬──┬───┘              
                {                                                                             │    │  │                  
                Rule:        "arista_interface_down",                                         │    │  │                  
                RawIncident: "test_device Ebra: 1417: %LINEPROTO-5-UPDOWN:              ┌─────┘    │  └─────┐            
                Line protocol on Interface Ethernet6/12/1, changed state to             │          │        │            
                down",                                                                  │          │        │            
                Parameters: {"hostname": test_device, "interface":                      │          │        │            
                "Ethernet6/12/1"},                                                      ▼          │        ▼            
                PreAudits:  "interface_check.pl --hostname 'test_device'      ┌───────────────────┐│┌───────────────────┐
                --interface 'Ethernet6/12/1'",                                │                   │││                   │
                            "capacity_link.pl --hostname 'test_device'        │      AUDITS       │││      AUDITS       │
                --interface 'Ethernet6/12/1'",                                │                   │││                   │
                Remediation:  "port_down_arista.py --hostname 'test_device'   └───────────────────┘│└───────────────────┘
                --interface 'Ethernet6/12/1'",                                 interface_check.pl  │   capacity_link.pl  
                Engine:      "syslog",                                                             │                     
                }                                                                                  │                     
                                                                                         ┌─────────▼─────────┐           
                                                                                         │                   │           
                                                                                         │    Remediation    │           
                                                                                         │                   │           
                                                                                         └───────────────────┘           
                                                                                          port_down_arista.py            

</pre>

# Deployment

All the "services" are designed to run detached: You create as many tailer and processors as needed (depending on your information source). The executor can also be instantiated to fit your load. 

Tailer and processors will handle thousands of messages per second with no issues; the executor will need to be scaled depending on the load of your jobs. Audits and remediations will run in the executor, on a heavy loaded environment, this might be your first point of scale.

<pre>
                                                                                                 ┌─────────────┐
                                                                                                 │             │
                                                                                            ┌───▶│   AUDITS    │
                                                                                            │    │             │
                                                                                            │    └─────────────┘
                                                                        ┌──────────────┐    │                   
                 ┌───────────────┐                                      │              │────┘                   
                 │               │                                 ┌────┤   EXECUTOR   │                        
      ┌──────────│   Processor   ├────┐                            │    │              │────┐                   
      │          │               │    │                            │    └──────────────┘    │                   
      │          └───────────────┘    │                            │                        │                   
      │                               │                            │                        │    ┌────────────┐ 
      │                               │       ┏━━━━━━━━━━━━━━━━┓   │                        │    │            │ 
      │                               │       ┃                ┃   │                        └───▶│    JOBS    │ 
      │          ┌───────────────┐    │       ┃                ◀───┘                             │            │ 
      │          │               │    └─────▶ ┃    RabbitMQ    ┃                                 └────────────┘ 
      │   ┌──────┤   Processor   ├───────────▶┃ QUEUE_INCIDENT ◀───┐                                            
      │   │      │               │    ┌─────▶ ┃                ┃   │                            ┌─────────────┐ 
      │   │      └───────────────┘    │       ┃                ┃   │                            │             │ 
      │   │                           │       ┗━━━━━━━━━━━━━━━━┛   │                       ┌───▶│   AUDITS    │ 
      │   │                           │                            │                       │    │             │ 
      │   │                           │                            │                       │    └─────────────┘ 
      │   │      ┌───────────────┐    │                            │   ┌──────────────┐    │                    
      │   │      │               │    │                            │   │              │────┘                    
      │   │      │   Processor   ├────┘                            └───┤   EXECUTOR   │                         
      │   │   ┌──│               │                                     │              │────┐                    
      │   │   │  └───────────────┘                                     └──────────────┘    │                    
      │   │   │                                                                            │                    
      │   │   │                                                                            │    ┌────────────┐  
      │   │   │                                                                            │    │            │  
      │   │   │                                                                            └───▶│    JOBS    │  
      │   │   │                                                                                 │            │  
 ┏━━━━▼━━━▼━━━▼━━━┓                                                                             └────────────┘  
 ┃                ┃                                                                                             
 ┃                ┃                                                                                             
 ┃    RabbitMQ    ┃◀──────────────────────────────────┐                                                         
 ┃   QUEUE_LOG    ┃◀───────────┐                      │                                                         
 ┃                ┃            │                      │                                                         
 ┃                ┃   ┌────────┴───────┐     ┌────────────────┐                                                 
 ┗━━━━━━━━━━━━━━━━┛   │                │     │                │                                                 
              ┌──────▶│     Tailer     │     │     Tailer     │                                                 
              │       │                │     │                │                                                 
              │       └───────▲────────┘     └────────▲───────┘                                                 
              │               │                       │                                                         
              │               │                       │                                                         
              │               │                       │                                                         
              │               │                       │                                                         
┌─────────────┴────┐ ┌────────┴─────────┐    ┌────────┴─────────┐                                               
│                  │ │                  │    │                  │                                               
│  Network Device  │ │  Network Device  │    │  Network Device  │                                               
│                  │ │                  │    │                  │                                               
└──────────────────┘ └──────────────────┘    └──────────────────┘                

</pre>

## Installation

You'll need to `go get`:

github.com/streadway/amqp
github.com/golang/glog
github.com/davecgh/go-spew/spew
gopkg.in/mcuadros/go-syslog.v2
github.com/hpcloud/tail

`go build` each of the services (or go run for testing):

Tailer and executor only need the binary and the config files for deployment (config.yaml). Processors will also need a rule definition (Example: rules.yaml)

For the expected behavior, you'll need to deploy the executor and a combination of tailers/processors (Example: FileTailer and the Syslog processor)

## License
GOAR is BSD licensed, as found in the LICENSE file.
