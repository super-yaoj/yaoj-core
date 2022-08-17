/*
Package workflow is for problem workflow manipulation.

# Workflow Graph

Workflow Graph is a directed acyclic graph (DAG) which describes how to
perform a single testcase's judgement.

Each node of the graph represents a processor, with its input files and
output files naming inbound and outbound respectively.

A key node is specially treated by (default) analyzer since it means the main
process of submission's testing. Currently the default analyzer adds all key
nodes' runtime duration and memory up to provide information to the testcase.

A directed edge goes either from one of the outbounds of the source (node) to
one of the inbounds of the destination (node), or from a field of a datagroup
to one of the inbounds of the destination (node).

Datagroups is where all data files are given from.

# Analyzer

An analyzer examines up all nodes' execution result and all generated files to
evaluate the whole process, and then returns a structured result.

# Builder

Builder provides a convenient way to create a workflow graph.

# Cache

To speed up submission testing, we cache the results of processors under
different level:

- Problem Wide (level 1): in memory cache, which is reset before each
submission.

- Global (level 2): cache data in file system, which provides cross-submission
cache sharing. This is extremely helpful when performing problem rejudge which
eliminates tons of compiling time.
*/
package workflow
