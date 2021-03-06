#!/usr/bin/env python

# Dirty hack to generate boilerplate for event proto-messages

import os
import sys
from subprocess import call

if len(sys.argv) < 2:
    print("usage: %s <proto-file>" % sys.argv[0])
    exit(1)
protoFile = sys.argv[1]

eventNames = []
with open(protoFile, "r") as fd:
    for line in fd.readlines():
        if not line.startswith("message"):
            continue
        eventNames.append(line.split(" ")[1])

h, t = os.path.split(protoFile)
outputFile = os.path.join(h, t.replace("proto", "gen.go"))
with open(outputFile, "w") as fd:
    fd.write("""// Code generated by hack/codegen-events.py. DO NOT EDIT.    
package events

import `github.com/golang/protobuf/proto`

type EventType = string

type Event interface {
    proto.Message
    Type() EventType
} 

""")

    # Generate the constant declarations
    consts = "const (\n"
    for eventName in eventNames:
        consts += "  %s EventType = \"%s\"\n" % ("Event" + eventName, eventName)
    consts += ")\n\n"
    fd.write(consts)

    # Generate functions
    for eventName in eventNames:
        fd.write("""func (m *%s) Type() EventType {
    return %s        
}
        
""" % (eventName, "Event" + eventName))


call(["gofmt", "-w", outputFile])


