import sqlalchemy as sa
import enum

from .db import (
    metadata,
    Column,
    CustomDateTime,
)

PRICE_HISTORY_PERIOD = 60 * 60 * 24 * 7 * 10e8  # 1 week

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

validator_prices = sa.Table(
    "validator_prices",
    metadata,
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id"), primary_key=True),
    Column("signal_id", sa.String, primary_key=True),
    Column("price_status", CustomPriceStatus),
    Column("price", sa.BigInteger),
    Column("timestamp", CustomDateTime),
)

delegator_signals = sa.Table(
    "delegator_signals",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("signal_id", sa.String, primary_key=True),
    Column("power", sa.BigInteger),
    Column("timestamp", CustomDateTime),
)

signal_total_powers = sa.Table(
    "signal_total_powers",
    metadata,
    Column("signal_id", sa.String, primary_key=True),
    Column("power", sa.BigInteger),
    sa.Index("ix_signal_total_powers_power", "power"),
)

prices = sa.Table(
    "prices",
    metadata,
    Column("signal_id", sa.String, primary_key=True),
    Column("price_status", CustomPriceStatus),
    Column("price", sa.BigInteger),
    Column("timestamp", CustomDateTime, primary_key=True),
)

reference_source_configs = sa.Table(
    "reference_source_configs",
    metadata,
    Column("ipfs_hash", sa.String),
    Column("version", sa.String),
    Column("timestamp", CustomDateTime, primary_key=True),
)

feeders = sa.Table(
    "feeders",
    metadata,
    Column("feeder_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("operator_address", sa.String, primary_key=True),
    sa.Index("idx_feeders_operator_address", "operator_address"),
)
