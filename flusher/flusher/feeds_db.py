import sqlalchemy as sa
import enum

from .db import (
    metadata,
    Column,
    CustomDateTime,
)

PRICE_HISTORY_PERIOD = 60 * 60 * 24 * 7 * 1e9  # 1 week


# Define the SignalPriceStatus Enum
class SignalPriceStatus(enum.Enum):
    Unspecified = 0
    Unsupported = 1
    Unavailable = 2
    Available = 3


class CustomSignalPriceStatus(sa.types.TypeDecorator):
    impl = sa.Enum(SignalPriceStatus)

    def process_bind_param(self, value, dialect):
        return SignalPriceStatus(value)


# Define the PriceStatus Enum
class PriceStatus(enum.Enum):
    Unspecified = 0
    UnknownSignalID = 1
    NotReady = 2
    Available = 3
    NotInCurrentFeeds = 4


class CustomPriceStatus(sa.types.TypeDecorator):
    impl = sa.Enum(PriceStatus)

    def process_bind_param(self, value, dialect):
        return PriceStatus(value)


# Define the FeedsEncoder Enum
class FeedsEncoder(enum.Enum):
    nil = 0
    fixed_point_abi = 1
    tick_abi = 2


class CustomFeedsEncoder(sa.types.TypeDecorator):
    impl = sa.Enum(FeedsEncoder)

    def process_bind_param(self, value, dialect):
        return FeedsEncoder(value)


feeds_signal_prices_txs = sa.Table(
    "feeds_signal_prices_txs",
    metadata,
    Column("transaction_id", sa.Integer, sa.ForeignKey("transactions.id"), primary_key=True),
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id"), primary_key=True),
    Column("feeder_id", sa.Integer, sa.ForeignKey("accounts.id")),
    Column("timestamp", CustomDateTime, index=True),
    sa.Index("ix_feeds_signal_prices_txs_validator_id_transaction_id", "validator_id", "transaction_id"),
    sa.Index("ix_feeds_signal_prices_txs_validator_id_timestamp", "validator_id", "timestamp"),
)

feeds_validator_prices = sa.Table(
    "feeds_validator_prices",
    metadata,
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id"), primary_key=True),
    Column("signal_id", sa.String, primary_key=True),
    Column("status", CustomSignalPriceStatus),
    Column("price", sa.BigInteger),
    Column("timestamp", CustomDateTime, index=True),
)

feeds_voter_signals = sa.Table(
    "feeds_voter_signals",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("signal_id", sa.String, primary_key=True),
    Column("power", sa.BigInteger),
    Column("timestamp", CustomDateTime, index=True),
)

feeds_signal_total_powers = sa.Table(
    "feeds_signal_total_powers",
    metadata,
    Column("signal_id", sa.String, primary_key=True),
    Column("power", sa.BigInteger, index=True),
)

feeds_historical_prices = sa.Table(
    "feeds_historical_prices",
    metadata,
    Column("signal_id", sa.String, primary_key=True),
    Column("timestamp", CustomDateTime, primary_key=True, index=True),
    Column("status", CustomPriceStatus),
    Column("price", sa.BigInteger),
)

feeds_reference_source_configs = sa.Table(
    "feeds_reference_source_configs",
    metadata,
    Column("timestamp", CustomDateTime, primary_key=True, index=True),
    Column("registry_ipfs_hash", sa.String),
    Column("registry_version", sa.String),
)

feeds_feeders = sa.Table(
    "feeds_feeders",
    metadata,
    Column("feeder_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("operator_address", sa.String, primary_key=True, index=True),
)
