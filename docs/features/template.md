# Name of the Feature

* Add diagrams wherever possible to explain things better
* Keep the end user, new community contributors in mind as target audience when you write this documentation; not an ovn expert
* Remove sections if they are not relevant for your feature
* Add new sections if your feature needs them
* This document should empower end users to start playing around with this feature and understand implementation details
* If you are looking for inspiration see admin network policies feature docs for a good example.

## Introduction

This section is important for producing high quality
user-focused documentation. Goal here is to highlight **what this
feature is all about**. This paragraph can be used to construct
release notes when we do releases.

Your Introduction should be one paragraph long. More detail
should go into the following sections.

## Motivation

This section is for explicitly listing the motivation, goals and
non-goals of this feature. Describe why the change is important and
the benefits to users. Goal here is to highlight **why is this
feature needed**. If there was an enhancement proposal written please
link that in this section so that readers can go through the user
stories there.

### User-Stories/Use-Cases

NOTE: If you had an enhancement proposal written for this feature
then copy the user stories over from there OR provide a link
to that enhancement proposal.

(What new user-stories/use-cases does this feature introduce?)

A user story should typically have a summary structured this way:

1. **As a** [user concerned by the story]
2. **I want** [goal of the story]
3. **so that** [reason for the story]

The “so that” part is optional if more details are provided in the description.
A story can also be supplemented with examples, diagrams, or additional notes.

e.g

Story 1: Deny traffic at a cluster level

As a cluster admin, I want to apply non-overridable deny rules to certain pod(s)
and(or) Namespace(s) that isolate the selected resources from all other cluster
internal traffic.

For Example: The admin wishes to protect a sensitive namespace by applying an
AdminNetworkPolicy which denies ingress from all other in-cluster resources
for all ports and protocols.

## How to enable this feature on an OVN-Kubernetes cluster?

Did you write new configs or knobs that need to be turned on
for this feature? If so educate end users on those details.
How can users disable this feature if they don't want it?

## Workflow Description

Explain how the user will use the feature. Be detailed and explicit.
Describe all of the actors, their roles, and the APIs or interfaces
involved. Define a starting state and then list the steps that the
user would need to go through to trigger the feature described in the
enhancement. **A detailed architectural diagram is a must to showcase
how this feature works in our CNI.** Remember that a picture can speak
a thousand words.


## Implementation Details

Use the following sub-sections to explain more implementation details in a top down fashion

### User facing API Changes

If any API changes were done, they must be outlined here and
look at the developer docs on how to generate API reference
documentation for this feature. Provide a link to that API
reference web page from this spot.
Also add details on where the CRD lives; example is it in
kubernetes or kubernetes-sigs or network-plumbing-working-group?

Provide a sample CRD/API.

### OVN-Kubernetes Implementation Details

Keep in mind goal is to educate users how this feature
was implemented in OVN-Kubernetes in this section.

What were the changes made to ovn-kubernetes control plane and
data plane to make this happen? Note differences if any for local
gateway versus shared gateway and default mode versus interconnect
mode. **A detailed OVN-Kubernetes networking topology diagram is a must
to showcase how this feature works in our CNI**. Remember that a picture
can speak a thousand words. Give details on how the above API is watched
and converted into OVN objects by OVN-Kubernetes. If there are changes
done on host-networking level like adding new routes they must
be highlighted.

#### OVN Constructs created in the databases

What were the OVN Objects used? Give snippets and details on
the various database objects that this feature creates in the
OVN NBDB and SBDB. Keep in mind goal is to educate users how this
feature was implemented using OVN in this section. You can go as deep
as you want like mentioning how the pipelines get executed
for your feature.

Provide sample nbctl commands to view the constructs and show the output.

#### OVS Flows generated

Provide details on specific changes done to br-ex, br-int or how
the relevant openflows generated by this feature look like.
Keep in mind goal is to educate users how this
feature was implemented using OVS in this section.

Provide sample ofctl commands to view the flows and show the output.

## Troubleshooting

Include details on how the end user can know if the
feature is correctly configured here? What are the different
ways to troubleshoot this feature?

* What metrics, alerts were added to warn users something has gone wrong?
* Debugging notes

## Best Practices

* Write best practices to be used/recommended for smooth or better
functioning of the feature if any; example "using namespaceSelectors"
will scale better than using "podSelectors" because flows generated will
be less for former than latter etc.

## Future Items

Add a bullet list of future plans if any

## Known Limitations

Add a bullet list of known limitations if any

## References

Provide links or other relevant details outside of
this documentation that you think end users should read.
If there were useful github discussions, public google docs
enhancement proposals etc they must be mentioned in this section