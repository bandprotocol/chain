import sqlalchemy as sa

from .db import metadata, Column, CustomDateTime

restake_vaults = sa.Table(
    "restake_vaults",
    metadata,
    Column("key", sa.String, primary_key=True),
    Column("is_active", sa.Boolean),
    Column("last_update", CustomDateTime, index=True),
)

restake_locks = sa.Table(
    "restake_locks",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("key", sa.String, sa.ForeignKey("restake_vaults.key"), primary_key=True),
    Column("power", sa.BigInteger),
    Column(
        "transaction_id", sa.Integer, sa.ForeignKey("transactions.id"), nullable=True
    ),
)

restake_historical_stakes = sa.Table(
    "restake_historical_stakes",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("timestamp", CustomDateTime, primary_key=True),
    Column("coins", sa.String),
)
