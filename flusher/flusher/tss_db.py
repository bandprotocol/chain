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


class TSSAccountStatus(enum.Enum):
    nil = 0
    active = 1
    inactive = 2
    jail = 3


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


class CustomTSSAccountStatus(sa.types.TypeDecorator):
    impl = sa.Enum(TSSAccountStatus)

    def process_bind_param(self, value, dialect):
        return TSSAccountStatus(value)


class CustomReplacementStatus(sa.types.TypeDecorator):
    impl = sa.Enum(ReplacementStatus)

    def process_bind_param(self, value, dialect):
        return ReplacementStatus(value)


signing_data = sa.Table(
    "signing_data",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("group_id", sa.Integer),
    Column("group_pub_key", CustomBase64),
    Column("msg", CustomBase64),
    Column("group_pub_nonce", CustomBase64),
    Column("signature", CustomBase64, nullable=True),
    Column("fee", sa.String),
    Column("status", CustomSigningStatus),
    Column("reason", sa.String, nullable=True),
    Column("created_height", sa.Integer, sa.ForeignKey("blocks.height"), nullable=True, index=True),
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), index=True),
    sa.Index("ix_group_id_group_pub_key_status", "group_id", "group_pub_key", "status"),
)

groups = sa.Table(
    "groups",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("size", sa.Integer),
    Column("threshold", sa.Integer),
    Column("dkg_context", CustomBase64),
    Column("pub_key", CustomBase64, nullable=True),
    Column("status", CustomGroupStatus),
    Column("fee", sa.String),
    # if zero set it to nil
    Column("latest_replacement_id", sa.integer, sa.ForeignKey("replacements.id"), nullable=True),
    Column("created_height", sa.Integer, index=True),
)

tss_accounts = sa.Table(
    "tss_accounts",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("status", CustomTSSAccountStatus),
    Column("since", CustomDateTime),
    Column("last_active", CustomDateTime),
)

members = sa.Table(
    "members",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("group_id", sa.Integer, sa.ForeignKey("groups.id"), primary_key=True),
    Column("account_id", sa.integer, sa.ForeignKey("tss_accounts.account_id")),
    Column("pub_key", CustomBase64, nullable=True),
    Column("is_malicious", sa.Boolean),
)

assigned_members = sa.Table(
    "assigned_members",
    metadata,
    Column("signing_id", sa.Integer, sa.ForeignKey("signing_data.id"), primary_key=True),
    Column("member_id", sa.Integer, primary_key=True),
    Column("group_id", sa.Integer, primary_key=True),
    Column("pub_d", CustomBase64),
    Column("pub_e", CustomBase64),
    Column("binding_factor", CustomBase64),
    Column("pub_nonce", CustomBase64),
    sa.ForeignKeyConstraint(["member_id", "group_id"], ["members.id", "members.group_id"]),
)

replacements = sa.Table(
    "replacements",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("signing_id", sa.Integer, sa.Foreignkey("signing_data.id")),
    Column("from_group_id", sa.Integer, sa.Foreignkey("groups.id")),
    Column("from_pub_key", CustomBase64),
    Column("to_group_id", sa.Integer, sa.ForeignKey("groups.id")),
    Column("to_pub_key", CustomBase64),
    Column("exec_time", CustomDateTime),
    Column("status", ReplacementStatus),
)
