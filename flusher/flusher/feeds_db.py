import sqlalchemy as sa

from .db import (
    metadata,
    Column,
    CustomDateTime,
)


validator_prices = sa.Table(
    "validator_prices",
    metadata,
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id"), primary_key=True),
    Column("signal_id", sa.String, primary_key=True),
    Column("price_status", sa.String),
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
)

prices = sa.Table(
    "prices",
    metadata,
    Column("signal_id", sa.String, primary_key=True),
    Column("price_status", sa.String),
    Column("price", sa.BigInteger),
    Column("timestamp", CustomDateTime),
)

reference_source_configs = sa.Table(
    "reference_source_configs",
    metadata,
    Column("ipfs_hash", sa.String,),
    Column("version", sa.String),
    Column("timestamp", CustomDateTime),
)

feeders = sa.Table(
    "feeders",
    metadata,
    Column("feeder_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("operator_address", sa.String, primary_key=True),
)
