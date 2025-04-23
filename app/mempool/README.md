# Mempool Package

The mempool package implements a transaction mempool for the Cosmos SDK blockchain. It provides a sophisticated transaction management system that organizes transactions into different lanes based on their types and priorities.

## Overview

The mempool is designed to handle transaction processing and block proposal preparation in a Cosmos SDK-based blockchain. It implements the `sdkmempool.Mempool` interface and uses a lane-based architecture to manage different types of transactions.

## Architecture

### Core Components

1. **Mempool**: The main structure that manages multiple transaction lanes and implements the core mempool interface.
2. **Lane**: A logical grouping of transactions with specific matching criteria and space allocation.
3. **Proposal**: Represents a block proposal under construction, managing transaction inclusion and space limits.
4. **BlockSpace**: Manages block space constraints including transaction bytes and gas limits.

### Key Features

- **Lane-based Organization**: Transactions are organized into different lanes based on their matching functions
- **Space Management**: Efficient management of block space and gas limits
- **Transaction Prioritization**: Support for different transaction priorities among and within lanes 
- **Thread Safety**: Built-in mutex protection for concurrent access
- **Error Recovery**: Panic recovery mechanisms for robust operation

## Usage

### Creating a Lane

A lane is created with specific matching criteria and space allocation. Here's an example of creating a lane for bank send transactions:

```go
bankSendLane := NewLane(
    logger,
    txEncoder,
    signerExtractor,
    "bankSend",                    // Lane name
    isBankSendTx,                  // Matching function
    math.LegacyMustNewDecFromStr("0.2"),  // Max transaction space ratio
    math.LegacyMustNewDecFromStr("0.3"),  // Max lane space ratio
    sdkmempool.DefaultPriorityMempool(),  // Underlying mempool implementation
    nil,                           // Lane limit check handler
)

// Example matching function for bank send transactions
func isBankSendTx(_ sdk.Context, tx sdk.Tx) bool {
    msgs := tx.GetMsgs()
    if len(msgs) == 0 {
        return false
    }
    for _, msg := range msgs {
        if _, ok := msg.(*banktypes.MsgSend); !ok {
            return false
        }
    }
    return true
}
```

Key parameters for lane creation:
- `logger`: Logger instance for lane operations
- `txEncoder`: Function to encode transactions
- `signerExtractor`: Adapter to extract signer information
- `name`: Unique identifier for the lane
- `matchFn`: Function to determine if a transaction belongs in this lane
- `maxTransactionSpace`: Maximum space ratio for individual transactions (relative to total block space) more details on [Space management](#space-management) 
- `maxLaneSpace`: Maximum space ratio for the entire lane (relative to total block space) more details on [Space management](#space-management)
- `laneMempool`: Underlying Cosmos SDK mempool implementation that handles transaction storage and ordering within the lane. This determines how transactions are stored and selected within the lane
- `handleLaneLimitCheck`: Optional callback function that is called when the lane exceeds its space limit. This can be used to implement inter-lane dependencies, such as blocking other lanes when a lane exceeds its limit.

#### Inter-Lane Dependencies

Lanes can be configured to interact with each other through the `handleLaneLimitCheck` callback. This is useful for implementing priority systems or dependencies between different types of transactions. For example, you might want to block certain lanes when a high-priority lane exceeds its limit:

```go
// Create a dependent lane that will be blocked when the dependency lane exceeds its limit
dependentLane := NewLane(
    logger,
    txEncoder,
    signerExtractor,
    "dependent",
    isOtherTx,
    math.LegacyMustNewDecFromStr("0.5"),
    math.LegacyMustNewDecFromStr("0.5"),
    sdkmempool.DefaultPriorityMempool(),
    nil,
)

// Create a dependency lane that controls the dependent lane
dependencyLane := NewLane(
    logger,
    txEncoder,
    signerExtractor,
    "dependency",
    isBankSendTx,
    math.LegacyMustNewDecFromStr("0.5"),
    math.LegacyMustNewDecFromStr("0.5"),
    sdkmempool.DefaultPriorityMempool(),
    func(isLaneLimitExceeded bool) {
        dependentLane.SetBlocked(isLaneLimitExceeded)
    },
)
```

In this example, when the dependency lane exceeds its space limit, it will block the dependent lane from processing transactions. This mechanism allows for sophisticated transaction prioritization and coordination between different types of transactions.

### Creating a Mempool

```go
mempool := NewMempool(
    logger,
    []*Lane{
        BankSendLane,
        DelegationLane,
        // Lane order is critical - first lane in the slice matches and processes the transaction first
    },
)

// set the mempool in Chain application
app.SetMempool(mempool)
```

Key parameters for mempool creation:
- `logger`: Logger instance for mempool operations
- `lanes`: Array of lane configurations, where each lane is responsible for a specific type of transaction
  - The order of lanes determines the priority of transaction types
  - The sum of all lane space ratios can exceed 100% as it represents the maximum potential allocation for each lane, not a strict partition of the block space

### Space Management

#### Space Allocation
Both `maxTransactionSpace` and `maxLaneSpace` are expressed as ratios of the total block space and are used specifically during proposal preparation. For example, a `maxTransactionSpace` of 0.2 means a single transaction can use up to 20% of the total block space in a proposal, while a `maxLaneSpace` of 0.3 means the entire lane can use up to 30% of the total block space in a proposal. These ratios are used to ensure fair distribution of block space among different transaction types during proposal construction.

#### Space Cap Behavior
- **Transaction Space Cap (Hard Cap)**: The `maxTransactionSpace` ratio enforces a strict limit on individual transaction sizes of each lane. For example, with a `maxTransactionSpace` of 0.2, a transaction requiring more than 20% of the total block space will be deleted from the lane.
- **Lane Space Cap (Soft Cap)**: The `maxLaneSpace` ratio serves as a guideline for space allocation during the first round of proposal construction. If a lane's `maxLaneSpace` is 0.3, it can still include one last transaction that would cause it to exceed this limit in the proposal, provided each individual transaction respects the `maxTransactionSpace` limit. For instance, a lane with a 0.3 `maxLaneSpace` could include two transactions each using 20% of the block space (totaling 40%) in the proposal, as long as both transactions individually respect the `maxTransactionSpace` limit.

#### Proposal Preparation
The lane space cap (`maxLaneSpace`) is only enforced during the first round of proposal preparation. In subsequent rounds, when filling the remaining block space, the lane cap is not considered, allowing lanes to potentially use more space than their initial allocation if space is available. This two-phase approach ensures both fair initial distribution and efficient use of remaining block space.

### Block Proposal Preparation

The mempool provides functionality to prepare block proposals by:
1. Filling proposals with transactions from each lane with `maxLaneSpace`
2. Filling remaining proposal space with transactions from each lane in the same order without `maxLaneSpace`
3. Handling transaction removal of the transactions that violate the `maxTransactionSpace` from all lanes

## Best Practices

1. Configure appropriate lane ratios based on your application's needs
2. Implement proper transaction matching functions for each lane
3. Always place the default lane as the last lane to ensure every transaction type can be matched

Note: Lane order is critical as it determines transaction processing priority. This is why the default lane should always be last, to ensure it only catches transactions that don't match any other lane.
