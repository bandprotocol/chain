import sqlalchemy as sa
import enum

from .db import (
    metadata,
    Column,
    CustomBase64,
    CustomDateTime,
)


class SigningStatus(enum.Enum):
    nil = 0
    waiting = 1
    success = 2
    expired = 3
    failed = 4


class GroupStatus(enum.Enum):
    nil = 0
    round1 = 1
    round2 = 2
    round3 = 3
    active = 4
    expired = 5
    failed = 6


class TSSStatus(enum.Enum):
    nil = 0
    active = 1
    paused = 2
    inactive = 3
    jail = 4


class ReplacementStatus(enum.Enum):
    nil = 0
    waiting = 1
    success = 2
    fallen = 3


class CustomSigningStatus(sa.types.TypeDecorator):
    impl = sa.Enum(SigningStatus)

    def process_bind_param(self, value, dialect):
        return SigningStatus(value)


class CustomGroupStatus(sa.types.TypeDecorator):
    impl = sa.Enum(GroupStatus)

    def process_bind_param(self, value, dialect):
        return GroupStatus(value)


class CustomTSSStatus(sa.types.TypeDecorator):
    impl = sa.Enum(TSSStatus)

    def process_bind_param(self, value, dialect):
        return TSSStatus(value)


class CustomReplacementStatus(sa.types.TypeDecorator):
    impl = sa.Enum(ReplacementStatus)

    def process_bind_param(self, value, dialect):
        return ReplacementStatus(value)


tss_signings = sa.Table(
    "tss_signings",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("tss_group_id", sa.Integer, sa.ForeignKey("tss_groups.id")),
    Column("group_pub_key", CustomBase64),
    Column("msg", CustomBase64),
    Column("group_pub_nonce", CustomBase64),
    Column("signature", CustomBase64, nullable=True),
    Column("fee", sa.String),
    Column("status", CustomSigningStatus),
    Column("reason", sa.String, nullable=True),
    Column(
        "created_height",
        sa.Integer,
        sa.ForeignKey("blocks.height"),
        nullable=True,
        index=True,
    ),
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), index=True),
    sa.Index(
        "ix_group_id_group_pub_key_status", "tss_group_id", "group_pub_key", "status"
    ),
)

tss_groups = sa.Table(
    "tss_groups",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("size", sa.Integer),
    Column("threshold", sa.Integer),
    Column("dkg_context", CustomBase64),
    Column("pub_key", CustomBase64, nullable=True),
    Column("status", CustomGroupStatus),
    Column("fee", sa.String),
    # if zero set it to nil
    Column(
        "latest_replacement_id",
        sa.Integer,
        sa.ForeignKey("tss_replacements.id"),
        nullable=True,
    ),
    Column("created_height", sa.Integer, index=True),
)

tss_statuses = sa.Table(
    "tss_statuses",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("status", CustomTSSStatus),
    Column("since", CustomDateTime),
    Column("last_active", CustomDateTime),
)

tss_group_members = sa.Table(
    "tss_group_members",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column(
        "tss_group_id", sa.Integer, sa.ForeignKey("tss_groups.id"), primary_key=True
    ),
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id")),
    Column("pub_key", CustomBase64, nullable=True),
    Column("is_malicious", sa.Boolean),
)

tss_assigned_members = sa.Table(
    "tss_assigned_members",
    metadata,
    Column(
        "tss_signing_id", sa.Integer, sa.ForeignKey("tss_signings.id"), primary_key=True
    ),
    Column(
        "tss_group_id",
        sa.Integer,
        primary_key=True,
    ),
    Column(
        "tss_group_member_id",
        sa.Integer,
        primary_key=True,
    ),
    Column("pub_d", CustomBase64),
    Column("pub_e", CustomBase64),
    Column("binding_factor", CustomBase64),
    Column("pub_nonce", CustomBase64),
    sa.ForeignKeyConstraint(
        ["tss_group_id", "tss_group_member_id"],
        ["tss_group_members.tss_group_id", "tss_group_members.id"],
    ),
)

tss_replacements = sa.Table(
    "tss_replacements",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("tss_signing_id", sa.Integer, sa.ForeignKey("tss_signings.id")),
    Column("from_group_id", sa.Integer, sa.ForeignKey("tss_groups.id")),
    Column("from_pub_key", CustomBase64),
    Column("to_group_id", sa.Integer, sa.ForeignKey("tss_groups.id")),
    Column("to_pub_key", CustomBase64),
    Column("exec_time", CustomDateTime),
    Column("status", CustomReplacementStatus),
)
