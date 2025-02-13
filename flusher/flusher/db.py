import base64 as b64
from datetime import datetime, date
import sqlalchemy as sa
import enum


class ResolveStatus(enum.Enum):
    Open = 0
    Success = 1
    Failure = 2
    Expired = 3


class ProposalStatus(enum.Enum):
    Nil = 0
    DepositPeriod = 1
    VotingPeriod = 2
    Passed = 3
    Rejected = 4
    Failed = 5
    Inactive = 6
    Cancelled = 7


class TSSEncoder(enum.Enum):
    Unspecified = 0
    Proto = 1
    FullABI = 2
    PartialABI = 3


class CustomResolveStatus(sa.types.TypeDecorator):
    impl = sa.Enum(ResolveStatus)

    def process_bind_param(self, value, dialect):
        return ResolveStatus(value)


class CustomProposalStatus(sa.types.TypeDecorator):
    impl = sa.Enum(ProposalStatus)

    def process_bind_param(self, value, dialect):
        return ProposalStatus(value)


class CustomTSSEncoder(sa.types.TypeDecorator):
    impl = sa.Enum(TSSEncoder)

    def process_bind_param(self, value, dialect):
        return TSSEncoder(value)


class CustomDateTime(sa.types.TypeDecorator):
    """Custom DateTime type that accepts Python nanosecond epoch int."""

    impl = sa.DateTime

    def process_bind_param(self, value, dialect):
        return datetime.fromtimestamp(value / 1e9) if value != None else None


class CustomBase64(sa.types.TypeDecorator):
    """Custom LargeBinary type that accepts base64-encoded string."""

    impl = sa.LargeBinary

    def process_bind_param(self, value, dialect):
        if value is None:
            return value
        return b64.decodebytes(value.encode())


class CustomDate(sa.types.TypeDecorator):
    """Custom Date type that accepts Python nanosecond epoch int."""

    impl = sa.Date

    def process_bind_param(self, value, dialect):
        dt = datetime.fromtimestamp(value / 1e9)
        return date(dt.year, dt.month, dt.day)


def Column(*args, **kwargs):
    """Forward into SQLAlchemy's Column construct, but with 'nullable' default to False."""
    if "nullable" not in kwargs:
        kwargs["nullable"] = False
    return sa.Column(*args, **kwargs)


metadata = sa.MetaData()


tracking = sa.Table(
    "tracking",
    metadata,
    Column("chain_id", sa.String, primary_key=True),
    Column("topic", sa.String),
    Column("kafka_offset", sa.BigInteger),
    Column("replay_topic", sa.String),
    Column("replay_offset", sa.Integer),
)


blocks = sa.Table(
    "blocks",
    metadata,
    Column("height", sa.Integer, primary_key=True),
    Column("timestamp", CustomDateTime, index=True),
    Column("proposer", sa.String, sa.ForeignKey("validators.consensus_address")),
    Column("hash", CustomBase64),
    Column("inflation", sa.Float),
    Column("supply", sa.String),  # uband suffix
)

transactions = sa.Table(
    "transactions",
    metadata,
    Column("id", sa.Integer, sa.Sequence("seq_transaction_id"), unique=True),
    Column("hash", CustomBase64, primary_key=True),
    Column("block_height", sa.Integer, sa.ForeignKey("blocks.height"), index=True),
    Column("gas_used", sa.Integer),
    Column("gas_limit", sa.Integer),
    Column("gas_fee", sa.String),  # uband suffix
    Column("err_msg", sa.String, nullable=True),
    Column("success", sa.Boolean),
    Column("sender", sa.String),
    Column("memo", sa.String),
    Column("messages", sa.JSON),
    Column("fee_payer", sa.String, nullable=True),
)

accounts = sa.Table(
    "accounts",
    metadata,
    Column("id", sa.Integer, sa.Sequence("seq_account_id"), unique=True),
    Column("address", sa.String, primary_key=True),
    Column("balance", sa.String),
)

interchain_accounts = sa.Table(
    "interchain_accounts",
    metadata,
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("connection_id", sa.String, sa.ForeignKey("connections.connection_id")),
    Column("counterparty_port", sa.String),
    Column("counterparty_address", sa.String),
)

data_sources = sa.Table(
    "data_sources",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("name", sa.String),
    Column("description", sa.String),
    Column("owner", sa.String),
    Column("executable", CustomBase64),
    Column("treasury", sa.String),
    Column("fee", sa.String),
    Column("transaction_id", sa.Integer, sa.ForeignKey("transactions.id"), nullable=True),
    Column("accumulated_revenue", sa.BigInteger),
    Column("last_request", CustomDateTime, nullable=True),
)

oracle_scripts = sa.Table(
    "oracle_scripts",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("name", sa.String),
    Column("description", sa.String),
    Column("owner", sa.String),
    Column("schema", sa.String),
    Column("codehash", sa.String),
    Column("source_code_url", sa.String),
    Column("transaction_id", sa.Integer, sa.ForeignKey("transactions.id"), nullable=True),
    Column("version", sa.Integer, nullable=True),
    Column("last_request", CustomDateTime, nullable=True),
)

requests = sa.Table(
    "requests",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("transaction_id", sa.Integer, sa.ForeignKey("transactions.id")),
    Column("oracle_script_id", sa.Integer, sa.ForeignKey("oracle_scripts.id")),
    Column("calldata", CustomBase64),
    Column("ask_count", sa.Integer),
    Column("min_count", sa.Integer),
    Column("fee_limit", sa.String),
    Column("prepare_gas", sa.Integer),
    Column("prepare_gas_used", sa.Integer, default=0),
    Column("execute_gas", sa.Integer),
    Column("execute_gas_used", sa.Integer, default=0),
    Column("tss_encoder", CustomTSSEncoder),
    Column("sender", sa.String, nullable=True),
    Column("client_id", sa.String),
    Column("request_time", sa.Integer, nullable=True, index=True),
    Column("resolve_status", CustomResolveStatus),
    Column("resolve_time", sa.Integer, nullable=True),
    Column("resolve_height", sa.Integer, sa.ForeignKey("blocks.height"), nullable=True, index=True),
    Column("reason", sa.String, nullable=True),
    Column("result", CustomBase64, nullable=True),
    Column("bandtss_signing_id", sa.Integer, sa.ForeignKey("bandtss_signings.id"), nullable=True),
    Column("bandtss_signing_error_codespace", sa.String, nullable=True),
    Column("bandtss_signing_error_code", sa.Integer, nullable=True),
    Column("total_fees", sa.String),
    Column("is_ibc", sa.Boolean),
    sa.Index(
        "ix_requests_oracle_script_id_id",
        "oracle_script_id",
        "id",
        postgresql_include=["transaction_id", "min_count", "ask_count", "resolve_status", "request_time"],
    ),
    sa.Index(
        "ix_requests_oracle_script_id_resolve_status_request_time",
        "oracle_script_id",
        "resolve_status",
        "request_time",
    ),
)

raw_requests = sa.Table(
    "raw_requests",
    metadata,
    Column("request_id", sa.Integer, sa.ForeignKey("requests.id"), primary_key=True),
    Column("external_id", sa.BigInteger, primary_key=True),
    Column("data_source_id", sa.Integer, sa.ForeignKey("data_sources.id"), index=True),
    Column("fee", sa.BigInteger),
    Column("calldata", CustomBase64),
    sa.Index("ix_raw_requests_data_source_id_request_id", "data_source_id", "request_id"),
)

val_requests = sa.Table(
    "val_requests",
    metadata,
    Column("request_id", sa.Integer, sa.ForeignKey("requests.id"), primary_key=True),
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id"), primary_key=True),
)

reports = sa.Table(
    "reports",
    metadata,
    Column("request_id", sa.Integer, sa.ForeignKey("requests.id"), primary_key=True),
    Column("transaction_id", sa.Integer, sa.ForeignKey("transactions.id")),
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id"), primary_key=True),
    Column("reporter_id", sa.Integer, sa.ForeignKey("accounts.id")),
    sa.Index("ix_reports_validator_id_request_id", "validator_id", "request_id"),
)

raw_reports = sa.Table(
    "raw_reports",
    metadata,
    Column("request_id", sa.Integer, primary_key=True),
    Column("validator_id", sa.Integer, primary_key=True),
    Column("external_id", sa.BigInteger, primary_key=True),
    Column("data", CustomBase64),
    Column("exit_code", sa.BigInteger),
    sa.ForeignKeyConstraint(["request_id", "validator_id"], ["reports.request_id", "reports.validator_id"]),
    sa.ForeignKeyConstraint(["request_id", "external_id"], ["raw_requests.request_id", "raw_requests.external_id"]),
)

validators = sa.Table(
    "validators",
    metadata,
    Column("id", sa.Integer, sa.Sequence("seq_validator_id"), unique=True),
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), unique=True),
    Column("operator_address", sa.String, primary_key=True),
    Column("consensus_address", sa.String, unique=True),
    Column("consensus_pubkey", sa.String),
    Column("moniker", sa.String),
    Column("identity", sa.String),
    Column("website", sa.String),
    Column("details", sa.String),
    Column("commission_rate", sa.String),
    Column("commission_max_rate", sa.String),
    Column("commission_max_change", sa.String),
    Column("min_self_delegation", sa.String),
    Column("jailed", sa.Boolean),
    Column("tokens", sa.DECIMAL),
    Column("delegator_shares", sa.DECIMAL),
    Column("current_reward", sa.DECIMAL),
    Column("current_ratio", sa.DECIMAL),
    Column("status", sa.Boolean, default=False),
    Column("status_since", CustomDateTime, default=0),
    Column("accumulated_commission", sa.String),  # uband suffix
)

delegations = sa.Table(
    "delegations",
    metadata,
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id"), primary_key=True),
    Column("delegator_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("shares", sa.DECIMAL),
    Column("last_ratio", sa.DECIMAL),
)

validator_votes = sa.Table(
    "validator_votes",
    metadata,
    Column("block_height", sa.Integer, sa.ForeignKey("blocks.height"), primary_key=True),
    Column("consensus_address", sa.String, sa.ForeignKey("validators.consensus_address"), primary_key=True),
    Column("voted", sa.Boolean),
)

unbonding_delegations = sa.Table(
    "unbonding_delegations",
    metadata,
    Column("delegator_id", sa.Integer, sa.ForeignKey("accounts.id")),
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id")),
    Column("creation_height", sa.Integer, sa.ForeignKey("blocks.height"), nullable=True),
    Column("completion_time", CustomDateTime),
    Column("amount", sa.DECIMAL),
)

redelegations = sa.Table(
    "redelegations",
    metadata,
    Column("delegator_id", sa.Integer, sa.ForeignKey("accounts.id")),
    Column("validator_src_id", sa.Integer, sa.ForeignKey("validators.id")),
    Column("validator_dst_id", sa.Integer, sa.ForeignKey("validators.id")),
    Column("completion_time", CustomDateTime),
    Column("amount", sa.DECIMAL),
)

account_transactions = sa.Table(
    "account_transactions",
    metadata,
    Column("transaction_id", sa.Integer, sa.ForeignKey("transactions.id"), primary_key=True),
    Column("account_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    sa.Index("ix_account_transactions_account_id_transaction_id", "account_id", "transaction_id"),
)

proposals = sa.Table(
    "proposals",
    metadata,
    Column("id", sa.Integer, primary_key=True),
    Column("proposer_id", sa.Integer, sa.ForeignKey("accounts.id"), nullable=True),
    Column("type", sa.String),
    Column("title", sa.String),
    Column("description", sa.String),
    Column("proposal_route", sa.String),
    Column("status", CustomProposalStatus),
    Column("submit_time", CustomDateTime, nullable=True),
    Column("deposit_end_time", CustomDateTime, nullable=True),
    Column("total_deposit", sa.String),  # uband suffix
    Column("voting_time", CustomDateTime, nullable=True),
    Column("voting_end_time", CustomDateTime, nullable=True),
    Column("content", sa.JSON, nullable=True),
    Column("total_bonded_tokens", sa.BigInteger, nullable=True),
    Column("yes_vote", sa.BigInteger, nullable=True),
    Column("no_vote", sa.BigInteger, nullable=True),
    Column("no_with_veto_vote", sa.BigInteger, nullable=True),
    Column("abstain_vote", sa.BigInteger, nullable=True),
)

deposits = sa.Table(
    "deposits",
    metadata,
    Column("proposal_id", sa.Integer, sa.ForeignKey("proposals.id"), primary_key=True),
    Column("depositor_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("amount", sa.String),  # uband suffix
    Column("tx_id", sa.Integer, sa.ForeignKey("transactions.id"), nullable=True),
)

votes = sa.Table(
    "votes",
    metadata,
    Column("proposal_id", sa.Integer, sa.ForeignKey("proposals.id"), primary_key=True),
    Column("voter_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("yes", sa.DECIMAL),
    Column("abstain", sa.DECIMAL),
    Column("no", sa.DECIMAL),
    Column("no_with_veto", sa.DECIMAL),
    Column("tx_id", sa.Integer, sa.ForeignKey("transactions.id"), nullable=True),
)

historical_bonded_token_on_validators = sa.Table(
    "historical_bonded_token_on_validators",
    metadata,
    Column("validator_id", sa.Integer, sa.ForeignKey("validators.id"), primary_key=True),
    Column("bonded_tokens", sa.DECIMAL),
    Column("timestamp", CustomDateTime, primary_key=True),
)

reporters = sa.Table(
    "reporters",
    metadata,
    Column("reporter_id", sa.Integer, sa.ForeignKey("accounts.id"), primary_key=True),
    Column("operator_address", sa.String, primary_key=True),
)

related_data_source_oracle_scripts = sa.Table(
    "related_data_source_oracle_scripts",
    metadata,
    Column("data_source_id", sa.Integer, sa.ForeignKey("data_sources.id"), primary_key=True),
    Column("oracle_script_id", sa.Integer, sa.ForeignKey("oracle_scripts.id"), primary_key=True),
)

historical_oracle_statuses = sa.Table(
    "historical_oracle_statuses",
    metadata,
    Column("operator_address", sa.String, primary_key=True),
    Column("status", sa.Boolean),
    Column("timestamp", CustomDateTime, primary_key=True),
)

data_source_requests = sa.Table(
    "data_source_requests",
    metadata,
    Column("data_source_id", sa.Integer, sa.ForeignKey("data_sources.id"), primary_key=True),
    Column("count", sa.Integer),
)

data_source_requests_per_days = sa.Table(
    "data_source_requests_per_days",
    metadata,
    Column("date", CustomDate, primary_key=True),
    Column("data_source_id", sa.Integer, sa.ForeignKey("data_sources.id"), primary_key=True),
    Column("count", sa.Integer),
)

oracle_script_requests = sa.Table(
    "oracle_script_requests",
    metadata,
    Column("oracle_script_id", sa.Integer, sa.ForeignKey("oracle_scripts.id"), primary_key=True),
    Column("count", sa.Integer),
)

oracle_script_requests_per_days = sa.Table(
    "oracle_script_requests_per_days",
    metadata,
    Column("date", CustomDate, primary_key=True),
    Column("oracle_script_id", sa.Integer, sa.ForeignKey("oracle_scripts.id"), primary_key=True),
    Column("count", sa.Integer),
)

request_count_per_days = sa.Table(
    "request_count_per_days",
    metadata,
    Column("date", CustomDate, primary_key=True),
    Column("count", sa.Integer),
)

incoming_packets = sa.Table(
    "incoming_packets",
    metadata,
    Column("block_height", sa.Integer, sa.ForeignKey("blocks.height"), index=True),
    Column("src_channel", sa.String),
    Column("src_port", sa.String),
    Column("sequence", sa.Integer, primary_key=True),
    Column("dst_channel", sa.String, primary_key=True),
    Column("dst_port", sa.String, primary_key=True),
    Column("tx_id", sa.Integer, sa.ForeignKey("transactions.id"), nullable=True),
    Column("type", sa.String, nullable=True),
    Column("data", sa.JSON, nullable=True),
    Column("acknowledgement", sa.JSON, nullable=True),
    sa.ForeignKeyConstraint(["dst_port", "dst_channel"], ["channels.port", "channels.channel"]),
)

outgoing_packets = sa.Table(
    "outgoing_packets",
    metadata,
    Column("block_height", sa.Integer, sa.ForeignKey("blocks.height"), index=True),
    Column("src_channel", sa.String, primary_key=True),
    Column("src_port", sa.String, primary_key=True),
    Column("sequence", sa.Integer, primary_key=True),
    Column("dst_channel", sa.String),
    Column("dst_port", sa.String),
    Column("tx_id", sa.Integer, sa.ForeignKey("transactions.id"), nullable=True),
    Column("type", sa.String, nullable=True),
    Column("data", sa.JSON, nullable=True),
    Column("acknowledgement", sa.JSON, nullable=True),
    sa.ForeignKeyConstraint(["src_port", "src_channel"], ["channels.port", "channels.channel"]),
)

counterparty_chains = sa.Table("counterparty_chains", metadata, Column("chain_id", sa.String, primary_key=True))

connections = sa.Table(
    "connections",
    metadata,
    Column("counterparty_chain_id", sa.String, sa.ForeignKey("counterparty_chains.chain_id"), primary_key=True),
    Column("counterparty_connection_id", sa.String),
    Column("client_id", sa.String),
    Column("counterparty_client_id", sa.String),
    Column("connection_id", sa.String, primary_key=True, unique=True),
)

channels = sa.Table(
    "channels",
    metadata,
    Column("connection_id", sa.String, sa.ForeignKey("connections.connection_id")),
    Column("port", sa.String, primary_key=True),
    Column("counterparty_port", sa.String),
    Column("channel", sa.String, primary_key=True),
    Column("counterparty_channel", sa.String),
    Column("state", sa.Integer),
    Column("order", sa.String),
    Column("last_update", CustomDateTime, index=True),
)

relayer_tx_stat_days = sa.Table(
    "relayer_tx_stat_days",
    metadata,
    Column("date", CustomDate, primary_key=True),
    Column("port", sa.String, primary_key=True),
    Column("channel", sa.String, primary_key=True),
    Column("address", sa.String, primary_key=True),
    Column("ibc_received_txs", sa.Integer),
    Column("last_update_at", CustomDateTime),
    sa.ForeignKeyConstraint(["port", "channel"], ["channels.port", "channels.channel"]),
)
