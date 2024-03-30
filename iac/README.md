# Foyle Support Infrastructure 

This directory contains the infrastructure to support development of the Foyle project. This largely consists of

* An artifact registry for publishing docker images


## Why Pulumi

I decided to give Pulumi a try because Anthos Config management is too expensive to leave it up and running.
With Pulumi we can just run the CLI each time we want to make a change.