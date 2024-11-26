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


class CustomSigningStatus(sa.types.TypeDecorator):
    impl = sa.Enum(SigningStatus)

    def process_bind_param(self, value, dialect):
        return SigningStatus(value)


class CustomGroupStatus(sa.types.TypeDecorator):
    impl = sa.Enum(GroupStatus)

    def process_bind_param(self, value, dialect):
        return GroupStatus(value)
    
tss_signings = sa.Table(
    "tss_signings",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("tss_group_id", sa.Integer, sa.ForeignKey("tss_groups.id")),
    Column("current_attempt", sa.Integer),
    Column("originator", CustomBase64),
    Column("message", CustomBase64),
    Column("group_pub_key", CustomBase64),
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

tss_signing_contents = sa.Table(
    "tss_signing_contents",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("content_info", CustomBase64),
    Column("originator_info", CustomBase64),
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

tss_assigned_members = sa.Table(
    "tss_assigned_members",
    metadata,
    Column(
        "tss_signing_id", sa.Integer, sa.ForeignKey("tss_signings.id"), primary_key=True
    ),
    Column("tss_signing_attempt", sa.Integer, primary_key=True),
    Column("tss_member_id", sa.Integer, primary_key=True),
    Column("tss_group_id", sa.Integer, sa.ForeignKey("tss_groups.id")),
    Column("pub_d", CustomBase64),
    Column("pub_e", CustomBase64),
    Column("binding_factor", CustomBase64),
    Column("pub_nonce", CustomBase64),
    Column("signature", CustomBase64, nullable=True),
    Column(
        "submitted_height", sa.Integer, sa.ForeignKey("blocks.height"), nullable=True
    ),
    sa.ForeignKeyConstraint(
        ["tss_group_id", "tss_member_id"],
        ["tss_members.tss_group_id", "tss_members.id"],
    ),
)
