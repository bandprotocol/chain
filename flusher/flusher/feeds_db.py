import sqlalchemy as sa
import enum

from .db import (
    metadata,
    Column,
    CustomDateTime,
)

PRICE_HISTORY_PERIOD = 60 * 60 * 24 * 7 * 1e9  # 1 week

# Define the PriceStatus Enum
class PriceStatus(enum.Enum):
    Unspecified = 0
    Unsupported = 1
    Unavailable = 2
    Available = 3

class CustomPriceStatus(sa.types.TypeDecorator):
    impl = sa.Enum(PriceStatus)

    def process_bind_param(self, value, dialect):
        return PriceStatus(value)

signal_prices_txs = sa.Table(
    "signal_prices_txs",
    metadata,
    Column("transaction_id", sa.Integer, sa.ForeignKey("transactions.id"), primary_key=True),
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id"), primary_key=True),
    Column("feeder_id", sa.Integer, sa.ForeignKey("accounts.id")),
    Column("timestamp", CustomDateTime, index=True),
    sa.Index("ix_signal_prices_txs_validator_id_transaction_id", "validator_id", "transaction_id"),
    sa.Index("ix_validator_id_timestamp", "validator_id", "timestamp"),
)

validator_prices = sa.Table(
    "validator_prices",
    metadata,
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id"), primary_key=True),
    Column("signal_id", sa.String, primary_key=True),
    Column("price_status", CustomPriceStatus),
    Column("price", sa.BigInteger),
    Column("timestamp", CustomDateTime, index=True),
)

voter_signals = sa.Table(
    "voter_signals",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("signal_id", sa.String, primary_key=True),
    Column("power", sa.BigInteger),
    Column("timestamp", CustomDateTime, index=True),
)

signal_total_powers = sa.Table(
    "signal_total_powers",
    metadata,
    Column("signal_id", sa.String, primary_key=True),
    Column("power", sa.BigInteger, index=True),
)

historical_prices = sa.Table(
    "historical_prices",
    metadata,
    Column("signal_id", sa.String, primary_key=True),
    Column("timestamp", CustomDateTime, primary_key=True, index=True),
    Column("price_status", CustomPriceStatus),
    Column("price", sa.BigInteger),
)

reference_source_configs = sa.Table(
    "reference_source_configs",
    metadata,
    Column("registry_ipfs_hash", sa.String),
    Column("registry_version", sa.String),
    Column("timestamp", CustomDateTime, primary_key=True, index=True),
)

feeders = sa.Table(
    "feeders",
    metadata,
    Column("feeder_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("operator_address", sa.String, primary_key=True, index=True),
)
