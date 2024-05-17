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
    Column("signal_id", sa.String, sa.ForeignKey("signal_total_powers.signal_id"), primary_key=True),
    Column("price", sa.BigInteger),
    Column("timestamp", CustomDateTime),
)

delegator_signals = sa.Table(
    "delegator_signals",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("signal_id", sa.String, sa.ForeignKey("signal_total_powers.signal_id"), primary_key=True),
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
    Column("signal_id", sa.String, sa.ForeignKey("signal_total_powers.signal_id"), primary_key=True),
    Column("price_status", sa.String),
    Column("price", sa.BigInteger),
    Column("timestamp", CustomDateTime),
)

price_services = sa.Table(
    "price_services",
    metadata,
    Column("hash", sa.String),
    Column("version", sa.String),
    Column("url", sa.String),
    Column("timestamp", CustomDateTime),
)
