import sqlalchemy as sa

from .db import (
    metadata,
    Column,
    CustomDateTime,
)


price_validators = sa.Table(
    "price_validators",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("symbol", sa.String, sa.ForeignKey("symbols.symbol"), primary_key=True),
    Column("price", sa.Integer),
    Column("timestamp", CustomDateTime),
)

delegator_signals = sa.Table(
    "delegator_signal",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("symbol", sa.String, sa.ForeignKey("symbols.symbol"), primary_key=True),
    Column("power", sa.Integer),
    Column("timestamp", CustomDateTime),
)

symbols = sa.Table(
    "symbols",
    metadata,
    Column("symbol", sa.String, primary_key=True),
    Column("power", sa.String),
    Column("interval", sa.Integer),
    Column("last_interval_update_timestamp", sa.Integer),
)

prices = sa.Table(
    "prices",
    metadata,
    Column("symbol", sa.String, sa.ForeignKey("symbols.symbol"), primary_key=True),
    Column("price", sa.Integer),
    Column("timestamp", CustomDateTime),
)
