import sqlalchemy as sa

from .db import (
    metadata,
    Column,
)

restake_vaults = sa.Table(
    "restake_vaults",
    metadata,
    Column("key", sa.String, primary_key=True),
    Column("vault_id", sa.Integer, sa.ForeignKey("accounts.id")),
    Column("is_active", sa.Boolean),
    Column("total_power", sa.BigInteger),
)

restake_locks = sa.Table(
    "restake_locks",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("key", sa.String, primary_key=True),
    Column("power", sa.BigInteger),
)

restake_stakes = sa.Table(
    "restake_stakes",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("coins", sa.String),
)
