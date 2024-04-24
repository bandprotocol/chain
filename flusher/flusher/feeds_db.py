import sqlalchemy as sa

from .db import (
    metadata,
    Column,
    CustomDateTime,
)


price_validators = sa.Table(
    "price_validators",
    metadata,
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id"), primary_key=True),
    Column("signal_id", sa.String, sa.ForeignKey("feeds.signal_id"), primary_key=True),
    Column("price", sa.Integer),
    Column("timestamp", CustomDateTime),
)

delegator_signals = sa.Table(
    "delegator_signals",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("signal_id", sa.String, sa.ForeignKey("feeds.signal_id"), primary_key=True),
    Column("power", sa.Integer),
    Column("timestamp", CustomDateTime),
)

feeds = sa.Table(
    "feeds",
    metadata,
    Column("signal_id", sa.String, primary_key=True),
    Column("power", sa.String),
    Column("interval", sa.Integer),
    Column("last_interval_update_timestamp", sa.Integer),
    Column("deviation_in_thousandth", sa.Integer),
)

prices = sa.Table(
    "prices",
    metadata,
    Column("symbol", sa.String, sa.ForeignKey("symbols.symbol"), primary_key=True),
    Column("price_option", sa.String),
    Column("price", sa.Integer),
    Column("timestamp", CustomDateTime),
)
