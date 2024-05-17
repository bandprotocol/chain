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
    fallen = 4


class GroupStatus(enum.Enum):
    nil = 0
    round1 = 1
    round2 = 2
    round3 = 3
    active = 4
    expired = 5
    fallen = 6


class ReplacementStatus(enum.Enum):
    nil = 0
    waiting_sign = 1
    waiting_replace = 2
    success = 3
    fallen = 4


class CustomSigningStatus(sa.types.TypeDecorator):
    impl = sa.Enum(SigningStatus)

    def process_bind_param(self, value, dialect):
        return SigningStatus(value)


class CustomGroupStatus(sa.types.TypeDecorator):
    impl = sa.Enum(GroupStatus)

    def process_bind_param(self, value, dialect):
        return GroupStatus(value)


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
    Column("status", CustomSigningStatus),
    Column("reason", sa.String, nullable=True),
    Column(
        "created_height",
        sa.Integer,
        sa.ForeignKey("blocks.height"),
        nullable=True,
        index=True,
    ),
    sa.Index(
        "ix_group_id_group_pub_key_status", "tss_group_id", "group_pub_key", "status"
    ),
)

band_tss_signings = sa.Table(
    "band_tss_signings",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("fee", sa.String),
    Column("requester_id", sa.Integer, sa.ForeignKey("accounts.id")),
    Column("current_group_signing_id", sa.Integer, sa.ForeignKey("tss_signings.id")),
    Column(
        "replacing_group_signing_id",
        sa.Integer,
        sa.ForeignKey("tss_signings.id"),
        nullable=True,
    ),
)

tss_groups = sa.Table(
    "tss_groups",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("size", sa.Integer),
    Column("threshold", sa.Integer),
    Column("pub_key", CustomBase64, nullable=True),
    Column("status", CustomGroupStatus),
    Column("dkg_context", CustomBase64),
    Column("module_owner", sa.String),
    Column("created_height", sa.Integer, index=True),
)

band_tss_groups = sa.Table(
    "band_tss_groups",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("current_group_id", sa.Integer, sa.ForeignKey("tss_groups.id")),
    Column("since", CustomDateTime),
)

tss_members = sa.Table(
    "tss_members",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column(
        "tss_group_id", sa.Integer, sa.ForeignKey("tss_groups.id"), primary_key=True
    ),
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id")),
    Column("pub_key", CustomBase64, nullable=True),
    Column("is_malicious", sa.Boolean),
    Column("is_active", sa.Boolean),
)

band_tss_members = sa.Table(
    "band_tss_members",
    metadata,
    Column("band_tss_groups_id", sa.ForeignKey("band_tss_groups.id"), primary_key=True),
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("is_active", sa.Boolean),
    Column("since", CustomDateTime),
    Column("last_active", CustomDateTime),
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
        sa.ForeignKey("tss_groups.id"),
        primary_key=True,
    ),
    Column(
        "tss_member_id",
        sa.Integer,
        primary_key=True,
    ),
    Column("pub_d", CustomBase64),
    Column("pub_e", CustomBase64),
    Column("binding_factor", CustomBase64),
    Column("pub_nonce", CustomBase64),
    sa.ForeignKeyConstraint(
        ["tss_group_id", "tss_member_id"],
        ["tss_members.tss_group_id", "tss_members.id"],
    ),
)

band_tss_replacements = sa.Table(
    "band_tss_replacements",
    metadata,
    Column(
        "tss_signing_id", sa.Integer, sa.ForeignKey("tss_signings.id"), primary_key=True
    ),
    Column("new_group_id", sa.Integer, sa.ForeignKey("tss_groups.id")),
    Column("new_pub_key", CustomBase64),
    Column("current_group_id", sa.Integer, sa.ForeignKey("tss_groups.id")),
    Column("current_pub_key", CustomBase64),
    Column("exec_time", CustomDateTime),
    Column("status", CustomReplacementStatus),
)
