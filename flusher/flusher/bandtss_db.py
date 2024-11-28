import enum

import sqlalchemy as sa

from .db import Column, CustomBase64, CustomDateTime, metadata


class GroupTransitionStatus(enum.Enum):
    nil = 0
    creating_group = 1
    waiting_sign = 2
    waiting_execution = 3
    success = 4
    expired = 5


class CustomGroupTransitionStatus(sa.types.TypeDecorator):
    impl = sa.Enum(GroupTransitionStatus)

    def process_bind_param(self, value, dialect):
        return GroupTransitionStatus(value)


bandtss_group_transitions = sa.Table(
    "bandtss_group_transitions",
    metadata,
    Column("proposal_id", sa.Integer, sa.ForeignKey("proposals.id"), primary_key=True),
    Column("tss_signing_id", sa.Integer, sa.ForeignKey("tss_signings.id"), nullable=True),
    Column(
        "current_tss_group_id",
        sa.Integer,
        sa.ForeignKey("tss_groups.id"),
        nullable=True,
    ),
    Column(
        "incoming_tss_group_id",
        sa.Integer,
        sa.ForeignKey("tss_groups.id"),
        nullable=True,
    ),
    Column("current_group_pub_key", CustomBase64, nullable=True),
    Column("incoming_group_pub_key", CustomBase64, nullable=True),
    Column("status", CustomGroupTransitionStatus),
    Column("exec_time", CustomDateTime),
    Column("is_force_transition", sa.Boolean),
    Column(
        "created_height",
        sa.Integer,
        sa.ForeignKey("blocks.height"),
        nullable=True,
        index=True,
    ),
    sa.Index(
        "ix_tss_signing_id_current_tss_group_id_incoming_tss_group_id",
        "tss_signing_id",
        "current_tss_group_id",
        "incoming_tss_group_id",
    ),
)

bandtss_historical_current_groups = sa.Table(
    "bandtss_historical_current_groups",
    metadata,
    Column(
        "proposal_id",
        sa.Integer,
        sa.ForeignKey("proposals.id"),
        nullable=True,
        index=True,
    ),
    Column(
        "current_tss_group_id",
        sa.Integer,
        sa.ForeignKey("tss_groups.id"),
        primary_key=True,
    ),
    Column(
        "transition_height",
        sa.Integer,
        sa.ForeignKey("blocks.height"),
        index=True,
        primary_key=True,
    ),
)

bandtss_members = sa.Table(
    "bandtss_members",
    metadata,
    Column("tss_group_id", sa.Integer, sa.ForeignKey("tss_groups.id"), primary_key=True),
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("is_active", sa.Boolean),
    Column("since", CustomDateTime, nullable=True),
)

bandtss_signings = sa.Table(
    "bandtss_signings",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id")),
    Column("fee_per_signer", sa.String),
    Column("total_fee", sa.String),
    Column(
        "current_group_tss_signing_id",
        sa.Integer,
        sa.ForeignKey("tss_signings.id"),
        nullable=True,
    ),
    Column(
        "incoming_group_tss_signing_id",
        sa.Integer,
        sa.ForeignKey("tss_signings.id"),
        nullable=True,
    ),
)
