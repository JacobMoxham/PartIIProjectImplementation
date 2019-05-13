# PartIIProjectImplementation
This project contains the code for *PAM*, a Policy Aware Middleware written in Go.

![](https://i.ibb.co/71GnJDR/PAM.png)

*PAM* supports three kinds of policy:
- Request policies: allow users to specify where they would prefer processing to
occur. They are attached to requests and specify:
  - The identity of the requester.
  - Whether all of the data needed to compute the query result is contained within
the query.
  - Whether it is preferred to perform the computation locally or remotely.
- Computation policies: are used to specify what processing a device is currently
willing to do. They dictate which HTTP requests can be handled and whether the
response will contain the raw data required to compute a result, or a full result.
- Data Policies: let users preserve their privacy by limiting what data specific re-
questers can view. They specify, for groups of requester identities, any columns of
database tables which should be excluded from queries, and any transforms which
should be applied to data before it is used in a query.

These are supported by three components as shown in the diagram above.

[The GoDocs for the project.](https://godoc.org/github.com/JacobMoxham/PartIIProjectImplementation/middleware)

It also contains code for several example usecases which can be run using `docker stack deploy`. These demonstrate its functionality.
