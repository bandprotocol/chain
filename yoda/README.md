# Yoda

## Introduction

Yoda is a program that is used by BandChain's validator nodes to automatically fulfill data for oracle requests.

Since a subset of validators who are selected for a data request must return the data they received from running the specified data source(s), each of them have to send a `MsgReportData` transaction to BandChain in order to fulfill their duty.

Although the transaction can be sent manually by the user, it is not convenient, and would be rather time-consuming. Furthermore, most data providers already have APIs that can be used to query data automatically by another software. Therefore, we have developed Yoda to help validators to automatically query data from data providers by executing data source script, then submit the result to fulfill the request.

For more details about Yoda, please follow this [link](https://docs.bandchain.org/node-validators/yoda)

## Installation

Please refer to [this documentation](https://docs.bandchain.org/node-validators/run-node/joining-mainnet/installation#step-5-setup-yoda) for the most up-to-date installation guide.
