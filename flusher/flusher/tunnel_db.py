import sqlalchemy as sa
import enum

from .db import (
    metadata,
    Column,
    CustomDateTime,
)


class TunnelEncoder(enum.Enum):
    nil = 0
    fixed_point_abi = 1
    tick_abi = 2


class DepositType(enum.Enum):
    nil = 0
    add = 1
    remove = 2


class CustomTunnelEncoder(sa.types.TypeDecorator):
    impl = sa.Enum(TunnelEncoder)

    def process_bind_param(self, value, dialect):
        return TunnelEncoder(value)


class CustomDepositType(sa.types.TypeDecorator):
    impl = sa.Enum(DepositType)

    def process_bind_param(self, value, dialect):
        return DepositType(value)


tunnels = sa.Table(
    "tunnels",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("sequence", sa.Integer),
    Column("route_type", sa.String),
    Column("route", sa.JSON),
    Column("encoder", CustomTunnelEncoder),
    Column("fee_payer_id", sa.Integer, sa.ForeignKey("accounts.id")),
    Column("total_deposit", sa.String),
    Column("status", sa.Boolean, index=True),
    Column("status_since", CustomDateTime, nullable=True),
    Column("last_interval", CustomDateTime),
    Column("creator_id", sa.Integer, sa.ForeignKey("accounts.id")),
    Column("created_at", CustomDateTime),
)

tunnel_historical_signal_deviations = sa.Table(
    "tunnel_historical_signal_deviations",
    metadata,
    Column("tunnel_id", sa.Integer, sa.ForeignKey("tunnels.id"), primary_key=True),
    Column("created_at", CustomDateTime, primary_key=True),
    Column("interval", sa.Integer, index=True),
    Column("signal_deviations", sa.JSON),
)

tunnel_deposits = sa.Table(
    "tunnel_deposits",
    metadata,
    Column("tunnel_id", sa.Integer, sa.ForeignKey("tunnels.id"), primary_key=True),
    Column("depositor_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("total_deposit", sa.String),
)

tunnel_historical_deposits = sa.Table(
    "tunnel_historical_deposits",
    metadata,
    Column(
        "transaction_id", sa.Integer, sa.ForeignKey("transactions.id"), primary_key=True
    ),
    Column("tunnel_id", sa.Integer, sa.ForeignKey("tunnels.id"), index=True),
    Column("depositor_id", sa.Integer, sa.ForeignKey("accounts.id"), index=True),
    Column("deposit_type", CustomDepositType),
    Column("amount", sa.String),
    Column("timestamp", CustomDateTime),
)

tunnel_packets = sa.Table(
    "tunnel_packets",
    metadata,
    Column("tunnel_id", sa.Integer, sa.ForeignKey("tunnels.id"), primary_key=True),
    Column("sequence", sa.Integer, primary_key=True),
    Column("packet_content_type", sa.String, index=True),
    Column("packet_content", sa.JSON),
    Column("base_fee", sa.String),
    Column("route_fee", sa.String),
    Column("created_at", CustomDateTime),
)

tunnel_packet_signal_prices = sa.Table(
    "tunnel_packet_signal_prices",
    metadata,
    Column("tunnel_id", sa.Integer, primary_key=True),
    Column("sequence", sa.Integer, primary_key=True),
    Column("signal_id", sa.String, primary_key=True),
    Column("price", sa.BigInteger),
    sa.ForeignKeyConstraint(
        ["tunnel_id", "sequence"],
        ["tunnel_packets.tunnel_id", "tunnel_packets.sequence"],
    ),
)
