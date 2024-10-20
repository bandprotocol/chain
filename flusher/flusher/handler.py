import base64 as b64
from datetime import datetime
from sqlalchemy import select
from sqlalchemy.dialects.postgresql import insert

from .db import (
    blocks,
    transactions,
    accounts,
    interchain_accounts,
    data_sources,
    oracle_scripts,
    requests,
    raw_requests,
    val_requests,
    reports,
    raw_reports,
    validators,
    delegations,
    validator_votes,
    unbonding_delegations,
    redelegations,
    account_transactions,
    proposals,
    deposits,
    votes,
    historical_bonded_token_on_validators,
    reporters,
    related_data_source_oracle_scripts,
    historical_oracle_statuses,
    data_source_requests,
    data_source_requests_per_days,
    oracle_script_requests,
    oracle_script_requests_per_days,
    request_count_per_days,
    incoming_packets,
    outgoing_packets,
    counterparty_chains,
    connections,
    channels,
    groups,
    group_members,
    group_policies,
    group_proposals,
    group_votes,
    relayer_tx_stat_days,
)

from .feeds_db import (
    PRICE_HISTORY_PERIOD,
    price_signals,
    validator_prices,
    delegator_signals,
    signal_total_powers,
    prices,
    reference_source_configs,
    feeders,
)


class Handler(object):
    def __init__(self, conn):
        self.conn = conn

    def get_transaction_id(self, tx_hash):
        return self.conn.execute(
            select([transactions.c.id]).where(transactions.c.hash == tx_hash)
        ).scalar()

    def get_transaction_sender(self, id):
        return self.conn.execute(
            select([transactions.c.sender]).where(transactions.c.id == id)
        ).scalar()

    def get_validator_id(self, val):
        return self.conn.execute(
            select([validators.c.id]).where(validators.c.operator_address == val)
        ).scalar()

    def get_account_id(self, address):
        if address is None:
            return None
        id = self.conn.execute(
            select([accounts.c.id]).where(accounts.c.address == address)
        ).scalar()
        if id is None:
            self.conn.execute(
                accounts.insert(), {"address": address, "balance": "0uband"}
            )
            return self.conn.execute(
                select([accounts.c.id]).where(accounts.c.address == address)
            ).scalar()
        return id

    def get_request_count(self, date):
        return self.conn.execute(
            select([request_count_per_days.c.count]).where(
                request_count_per_days.c.date == date
            )
        ).scalar()

    def get_oracle_script_requests_count_per_day(self, date, oracle_script_id):
        return self.conn.execute(
            select([oracle_script_requests_per_days.c.count]).where(
                (oracle_script_requests_per_days.c.date == date)
                & (
                    oracle_script_requests_per_days.c.oracle_script_id
                    == oracle_script_id
                )
            )
        ).scalar()

    def get_data_source_requests_count_per_day(self, date, data_source_id):
        return self.conn.execute(
            select([data_source_requests_per_days.c.count]).where(
                (data_source_requests_per_days.c.date == date)
                & (data_source_requests_per_days.c.data_source_id == data_source_id)
            )
        ).scalar()

    def get_data_source_id(self, id):
        return self.conn.execute(
            select([data_sources.c.id]).where(data_sources.c.id == id)
        ).scalar()

    def get_oracle_script_id(self, id):
        return self.conn.execute(
            select([oracle_scripts.c.id]).where(oracle_scripts.c.id == id)
        ).scalar()

    def get_group_id_from_policy_address(self, address):
        return self.conn.execute(
            select([group_policies.c.group_id]).where(
                group_policies.c.address == address
            )
        ).scalar()

    def get_ibc_received_txs(self, date, port, channel, address):
        msg = {"date": date, "port": port, "channel": channel, "address": address}
        condition = True
        for col in relayer_tx_stat_days.primary_key.columns.values():
            condition = (col == msg[col.name]) & condition

        return self.conn.execute(
            select([relayer_tx_stat_days.c.ibc_received_txs]).where(condition)
        ).scalar()

    def handle_new_block(self, msg):
        self.conn.execute(blocks.insert(), msg)

    def handle_new_transaction(self, msg):
        msg["fee_payer"] = (
            msg["fee_payer"] if "fee_payer" in msg and len(msg["fee_payer"]) else None
        )
        self.conn.execute(
            insert(transactions)
            .values(**msg)
            .on_conflict_do_update(constraint="transactions_pkey", set_=msg)
        )

    def handle_set_related_transaction(self, msg):
        tx_id = self.get_transaction_id(msg["hash"])
        related_tx_accounts = msg["related_accounts"]
        for account in related_tx_accounts:
            self.conn.execute(
                insert(account_transactions)
                .values(
                    {
                        "transaction_id": tx_id,
                        "account_id": self.get_account_id(account),
                    }
                )
                .on_conflict_do_nothing(constraint="account_transactions_pkey")
            )

    def handle_set_account(self, msg):
        if self.get_account_id(msg["address"]) is None:
            self.conn.execute(accounts.insert(), msg)
        else:
            condition = True
            for col in accounts.primary_key.columns.values():
                condition = (col == msg[col.name]) & condition
            self.conn.execute(accounts.update().where(condition).values(**msg))

    def handle_new_interchain_account(self, msg):
        msg["account_id"] = self.get_account_id(msg["address"])
        del msg["address"]
        self.conn.execute(
            insert(interchain_accounts)
            .values(**msg)
            .on_conflict_do_update(constraint="interchain_accounts_pkey", set_=msg)
        )

    def handle_new_data_source(self, msg):
        if msg["tx_hash"] is not None:
            msg["transaction_id"] = self.get_transaction_id(msg["tx_hash"])
        else:
            msg["transaction_id"] = None
        del msg["tx_hash"]
        msg["accumulated_revenue"] = 0
        self.conn.execute(data_sources.insert(), msg)
        self.init_data_source_request_count(msg["id"])

    def handle_new_group(self, msg):
        self.conn.execute(groups.insert(), msg)

    def handle_new_group_member(self, msg):
        msg["account_id"] = self.get_account_id(msg["address"])
        del msg["address"]
        self.conn.execute(group_members.insert(), msg)

    def handle_new_group_policy(self, msg):
        self.get_account_id(msg["address"])
        self.conn.execute(group_policies.insert(), msg)

    def handle_new_group_proposal(self, msg):
        msg["group_id"] = self.get_group_id_from_policy_address(
            msg["group_policy_address"]
        )
        self.conn.execute(group_proposals.insert(), msg)

    def handle_new_group_vote(self, msg):
        msg["voter_id"] = self.get_account_id(msg["voter_address"])
        del msg["voter_address"]
        self.conn.execute(group_votes.insert(), msg)

    def handle_update_group(self, msg):
        self.conn.execute(groups.update().where(groups.c.id == msg["id"]).values(**msg))

    def handle_remove_group_member(self, msg):
        account_id = self.get_account_id(msg["address"])
        self.conn.execute(
            group_members.delete().where(
                (group_members.c.group_id == msg["group_id"])
                & (group_members.c.account_id == account_id)
            )
        )

    def handle_remove_group_members_by_group_id(self, msg):
        self.conn.execute(
            group_members.delete().where(group_members.c.group_id == msg["group_id"])
        )

    def handle_update_group_policy(self, msg):
        self.conn.execute(
            group_policies.update()
            .where(group_policies.c.address == msg["address"])
            .values(**msg)
        )

    def handle_update_group_proposal(self, msg):
        msg["group_id"] = self.get_group_id_from_policy_address(
            msg["group_policy_address"]
        )
        self.conn.execute(
            group_proposals.update()
            .where(group_proposals.c.id == msg["id"])
            .values(**msg)
        )

    def handle_update_group_proposal_by_id(self, msg):
        self.conn.execute(
            group_proposals.update()
            .where(group_proposals.c.id == msg["id"])
            .values(**msg)
        )

    def handle_set_data_source(self, msg):
        msg["transaction_id"] = self.get_transaction_id(msg["tx_hash"])
        del msg["tx_hash"]
        condition = True
        for col in data_sources.primary_key.columns.values():
            condition = (col == msg[col.name]) & condition
        self.conn.execute(data_sources.update().where(condition).values(**msg))

    def handle_set_oracle_script(self, msg):
        if msg["tx_hash"] is not None:
            msg["transaction_id"] = self.get_transaction_id(msg["tx_hash"])
        else:
            msg["transaction_id"] = None
        del msg["tx_hash"]
        if self.get_oracle_script_id(msg["id"]) is None:
            self.conn.execute(oracle_scripts.insert(), msg)
            self.init_oracle_script_request_count(msg["id"])
        else:
            condition = True
            for col in oracle_scripts.primary_key.columns.values():
                condition = (col == msg[col.name]) & condition
            self.conn.execute(oracle_scripts.update().where(condition).values(**msg))

    def handle_new_request(self, msg):
        msg["transaction_id"] = self.get_transaction_id(msg["tx_hash"])
        del msg["tx_hash"]
        if "timestamp" in msg:
            self.handle_set_request_count_per_day({"date": msg["timestamp"]})
            self.handle_update_oracle_script_requests_count_per_day(
                {"date": msg["timestamp"], "oracle_script_id": msg["oracle_script_id"]}
            )
            self.update_oracle_script_last_request(
                msg["oracle_script_id"], msg["timestamp"]
            )
            del msg["timestamp"]
        self.conn.execute(requests.insert(), msg)
        self.increase_oracle_script_count(msg["oracle_script_id"])

    def handle_update_request(self, msg):
        if "tx_hash" in msg:
            msg["transaction_id"] = self.get_transaction_id(msg["tx_hash"])
            del msg["tx_hash"]
        condition = True
        for col in requests.primary_key.columns.values():
            condition = (col == msg[col.name]) & condition
        self.conn.execute(requests.update().where(condition).values(**msg))

    def handle_update_related_ds_os(self, msg):
        self.conn.execute(
            insert(related_data_source_oracle_scripts)
            .values(**msg)
            .on_conflict_do_nothing(
                constraint="related_data_source_oracle_scripts_pkey"
            )
        )

    def handle_new_raw_request(self, msg):
        self.increase_data_source_count(msg["data_source_id"])
        if "timestamp" in msg:
            self.handle_update_data_source_requests_count_per_day(
                {"date": msg["timestamp"], "data_source_id": msg["data_source_id"]}
            )
            self.update_data_source_last_request(
                msg["data_source_id"], msg["timestamp"]
            )
            del msg["timestamp"]
        self.handle_update_related_ds_os(
            {
                "oracle_script_id": self.conn.execute(
                    select([requests.c.oracle_script_id]).where(
                        requests.c.id == msg["request_id"]
                    )
                ).scalar(),
                "data_source_id": msg["data_source_id"],
            }
        )
        self.conn.execute(raw_requests.insert(), msg)
        self.increase_accumulated_revenue(msg["data_source_id"], msg["fee"])

    def increase_accumulated_revenue(self, id, fee):
        self.conn.execute(
            data_sources.update(data_sources.c.id == id).values(
                accumulated_revenue=data_sources.c.accumulated_revenue + fee
            )
        )

    def handle_new_val_request(self, msg):
        msg["validator_id"] = self.get_validator_id(msg["validator"])
        del msg["validator"]
        self.conn.execute(val_requests.insert(), msg)

    def handle_new_report(self, msg):
        if msg["tx_hash"] is not None:
            msg["transaction_id"] = self.get_transaction_id(msg["tx_hash"])
        del msg["tx_hash"]
        msg["validator_id"] = self.get_validator_id(msg["validator"])
        del msg["validator"]
        msg["reporter_id"] = self.get_account_id(msg["reporter"])
        del msg["reporter"]
        self.conn.execute(reports.insert(), msg)

    def handle_new_raw_report(self, msg):
        msg["validator_id"] = self.get_validator_id(msg["validator"])
        del msg["validator"]
        self.conn.execute(raw_reports.insert(), msg)

    def handle_set_validator(self, msg):
        last_update = msg["last_update"]
        del msg["last_update"]
        msg["account_id"] = self.get_account_id(msg["delegator_address"])
        del msg["delegator_address"]
        if self.get_validator_id(msg["operator_address"]) is None:
            self.conn.execute(validators.insert(), msg)
        else:
            condition = True
            for col in validators.primary_key.columns.values():
                condition = (col == msg[col.name]) & condition
            self.conn.execute(validators.update().where(condition).values(**msg))
        self.handle_new_historical_bonded_token_on_validator(
            {
                "validator_id": self.get_validator_id(msg["operator_address"]),
                "bonded_tokens": msg["tokens"],
                "timestamp": last_update,
            }
        )

    def handle_update_validator(self, msg):
        if "tokens" in msg and "last_update" in msg:
            self.handle_new_historical_bonded_token_on_validator(
                {
                    "validator_id": self.get_validator_id(msg["operator_address"]),
                    "bonded_tokens": msg["tokens"],
                    "timestamp": msg["last_update"],
                }
            )
            del msg["last_update"]
        self.conn.execute(
            validators.update()
            .where(validators.c.operator_address == msg["operator_address"])
            .values(**msg)
        )

    def handle_set_delegation(self, msg):
        msg["delegator_id"] = self.get_account_id(msg["delegator_address"])
        del msg["delegator_address"]
        msg["validator_id"] = self.get_validator_id(msg["operator_address"])
        del msg["operator_address"]
        self.conn.execute(
            insert(delegations)
            .values(**msg)
            .on_conflict_do_update(constraint="delegations_pkey", set_=msg)
        )

    def handle_update_delegation(self, msg):
        msg["delegator_id"] = self.get_account_id(msg["delegator_address"])
        del msg["delegator_address"]
        msg["validator_id"] = self.get_validator_id(msg["operator_address"])
        del msg["operator_address"]
        condition = True
        for col in delegations.primary_key.columns.values():
            condition = (col == msg[col.name]) & condition
        self.conn.execute(delegations.update().where(condition).values(**msg))

    def handle_remove_delegation(self, msg):
        msg["delegator_id"] = self.get_account_id(msg["delegator_address"])
        del msg["delegator_address"]
        msg["validator_id"] = self.get_validator_id(msg["operator_address"])
        del msg["operator_address"]
        condition = True
        for col in delegations.primary_key.columns.values():
            condition = (col == msg[col.name]) & condition
        self.conn.execute(delegations.delete().where(condition))

    def handle_new_validator_vote(self, msg):
        self.conn.execute(insert(validator_votes).values(**msg))

    def handle_new_unbonding_delegation(self, msg):
        msg["delegator_id"] = self.get_account_id(msg["delegator_address"])
        del msg["delegator_address"]
        msg["validator_id"] = self.get_validator_id(msg["operator_address"])
        del msg["operator_address"]
        self.conn.execute(insert(unbonding_delegations).values(**msg))

    def handle_remove_unbonding(self, msg):
        self.conn.execute(
            unbonding_delegations.delete().where(
                unbonding_delegations.c.completion_time <= msg["timestamp"]
            )
        )

    def handle_new_redelegation(self, msg):
        msg["delegator_id"] = self.get_account_id(msg["delegator_address"])
        del msg["delegator_address"]
        msg["validator_src_id"] = self.get_validator_id(msg["operator_src_address"])
        del msg["operator_src_address"]
        msg["validator_dst_id"] = self.get_validator_id(msg["operator_dst_address"])
        del msg["operator_dst_address"]
        self.conn.execute(insert(redelegations).values(**msg))

    def handle_remove_redelegation(self, msg):
        self.conn.execute(
            redelegations.delete().where(
                redelegations.c.completion_time <= msg["timestamp"]
            )
        )

    def handle_new_proposal(self, msg):
        msg["proposer_id"] = self.get_account_id(msg["proposer"])
        del msg["proposer"]
        self.conn.execute(proposals.insert(), msg)

    def handle_set_deposit(self, msg):
        msg["depositor_id"] = self.get_account_id(msg["depositor"])
        del msg["depositor"]
        msg["tx_id"] = self.get_transaction_id(msg["tx_hash"])
        del msg["tx_hash"]
        self.conn.execute(
            insert(deposits)
            .values(**msg)
            .on_conflict_do_update(constraint="deposits_pkey", set_=msg)
        )

    def handle_set_vote_weighted(self, msg):
        msg["voter_id"] = self.get_account_id(msg["voter"])
        del msg["voter"]
        msg["tx_id"] = self.get_transaction_id(msg["tx_hash"])
        del msg["tx_hash"]
        self.conn.execute(
            insert(votes)
            .values(**msg)
            .on_conflict_do_update(constraint="votes_pkey", set_=msg)
        )

    def handle_update_proposal(self, msg):
        condition = True
        for col in proposals.primary_key.columns.values():
            condition = (col == msg[col.name]) & condition
        self.conn.execute(proposals.update().where(condition).values(**msg))

    def handle_set_historical_bonded_token_on_validator(self, msg):
        msg["validator_id"] = self.get_validator_id(msg["operator_address"])
        del msg["operator_address"]
        self.conn.execute(
            insert(historical_bonded_token_on_validators)
            .values(**msg)
            .on_conflict_do_update(
                constraint="historical_bonded_token_on_validators_pkey", set_=msg
            )
        )

    def handle_set_reporter(self, msg):
        msg["operator_address"] = msg["validator"]
        del msg["validator"]
        msg["reporter_id"] = self.get_account_id(msg["reporter"])
        del msg["reporter"]
        self.conn.execute(
            insert(reporters)
            .values(msg)
            .on_conflict_do_nothing(constraint="reporters_pkey")
        )

    def handle_remove_reporter(self, msg):
        msg["operator_address"] = msg["validator"]
        del msg["validator"]
        msg["reporter_id"] = self.get_account_id(msg["reporter"])
        del msg["reporter"]
        condition = True
        for col in reporters.primary_key.columns.values():
            condition = (col == msg[col.name]) & condition
        self.conn.execute(reporters.delete().where(condition))

    def handle_set_historical_validator_status(self, msg):
        self.conn.execute(
            insert(historical_oracle_statuses)
            .values(**msg)
            .on_conflict_do_update(
                constraint="historical_oracle_statuses_pkey", set_=msg
            )
        )

    def init_data_source_request_count(self, id):
        self.conn.execute(
            insert(data_source_requests)
            .values({"data_source_id": id, "count": 0})
            .on_conflict_do_nothing(constraint="data_source_requests_pkey")
        )

    def increase_data_source_count(self, id):
        self.conn.execute(
            data_source_requests.update(
                data_source_requests.c.data_source_id == id
            ).values(count=data_source_requests.c.count + 1)
        )

    def init_oracle_script_request_count(self, id):
        self.conn.execute(
            insert(oracle_script_requests)
            .values({"oracle_script_id": id, "count": 0})
            .on_conflict_do_nothing(constraint="oracle_script_requests_pkey")
        )

    def handle_update_oracle_script_request(self, msg):
        condition = True
        for col in oracle_script_requests.primary_key.columns.values():
            condition = (col == msg[col.name]) & condition
        self.conn.execute(
            oracle_script_requests.update(condition).values(
                count=oracle_script_requests.c.count + 1
            )
        )

    def handle_set_request_count_per_day(self, msg):
        if self.get_request_count(msg["date"]) is None:
            msg["count"] = 1
            self.conn.execute(request_count_per_days.insert(), msg)
        else:
            condition = True
            for col in request_count_per_days.primary_key.columns.values():
                condition = (col == msg[col.name]) & condition
            self.conn.execute(
                request_count_per_days.update(condition).values(
                    count=request_count_per_days.c.count + 1
                )
            )

    def handle_update_oracle_script_requests_count_per_day(self, msg):
        if (
            self.get_oracle_script_requests_count_per_day(
                msg["date"], msg["oracle_script_id"]
            )
            is None
        ):
            msg["count"] = 1
            self.conn.execute(oracle_script_requests_per_days.insert(), msg)
        else:
            condition = True
            for col in oracle_script_requests_per_days.primary_key.columns.values():
                condition = (col == msg[col.name]) & condition
            self.conn.execute(
                oracle_script_requests_per_days.update(condition).values(
                    count=oracle_script_requests_per_days.c.count + 1
                )
            )

    def handle_update_data_source_requests_count_per_day(self, msg):
        if (
            self.get_data_source_requests_count_per_day(
                msg["date"], msg["data_source_id"]
            )
            is None
        ):
            msg["count"] = 1
            self.conn.execute(data_source_requests_per_days.insert(), msg)
        else:
            condition = True
            for col in data_source_requests_per_days.primary_key.columns.values():
                condition = (col == msg[col.name]) & condition
            self.conn.execute(
                data_source_requests_per_days.update(condition).values(
                    count=data_source_requests_per_days.c.count + 1
                )
            )

    def handle_new_incoming_packet(self, msg):
        self.update_last_update_channel(
            msg["dst_port"], msg["dst_channel"], msg["block_time"]
        )

        msg["tx_id"] = self.get_transaction_id(msg["hash"])
        del msg["hash"]

        msg["sender"] = self.get_transaction_sender(msg["tx_id"])
        self.handle_set_relayer_tx_stat_days(
            msg["dst_port"], msg["dst_channel"], msg["block_time"], msg["sender"]
        )
        del msg["block_time"]
        del msg["sender"]

        self.conn.execute(
            insert(incoming_packets)
            .values(**msg)
            .on_conflict_do_nothing(constraint="incoming_packets_pkey")
        )

    def handle_new_outgoing_packet(self, msg):
        self.update_last_update_channel(
            msg["src_port"], msg["src_channel"], msg["block_time"]
        )
        del msg["block_time"]

        msg["tx_id"] = self.get_transaction_id(msg["hash"])
        del msg["hash"]

        self.conn.execute(
            insert(outgoing_packets)
            .values(**msg)
            .on_conflict_do_nothing(constraint="outgoing_packets_pkey")
        )

    def handle_update_outgoing_packet(self, msg):
        self.update_last_update_channel(
            msg["src_port"], msg["src_channel"], msg["block_time"]
        )
        del msg["block_time"]

        condition = True
        for col in outgoing_packets.primary_key.columns.values():
            condition = (col == msg[col.name]) & condition
        self.conn.execute(outgoing_packets.update(condition).values(**msg))

    def increase_oracle_script_count(self, id):
        self.conn.execute(
            oracle_script_requests.update(
                oracle_script_requests.c.oracle_script_id == id
            ).values(count=oracle_script_requests.c.count + 1)
        )

    def update_oracle_script_last_request(self, id, timestamp):
        self.conn.execute(
            oracle_scripts.update(oracle_scripts.c.id == id).values(
                last_request=timestamp
            )
        )

    def update_data_source_last_request(self, id, timestamp):
        self.conn.execute(
            data_sources.update(data_sources.c.id == id).values(last_request=timestamp)
        )

    def handle_new_historical_bonded_token_on_validator(self, msg):
        self.conn.execute(
            insert(historical_bonded_token_on_validators)
            .values(**msg)
            .on_conflict_do_update(
                constraint="historical_bonded_token_on_validators_pkey", set_=msg
            )
        )

    def handle_set_counterparty_chain(self, msg):
        self.conn.execute(
            insert(counterparty_chains)
            .values(**msg)
            .on_conflict_do_update(constraint="counterparty_chains_pkey", set_=msg)
        )

    def handle_set_connection(self, msg):
        self.conn.execute(
            insert(connections)
            .values(**msg)
            .on_conflict_do_update(constraint="connections_pkey", set_=msg)
        )

    def handle_set_channel(self, msg):
        self.conn.execute(
            insert(channels)
            .values(**msg)
            .on_conflict_do_update(constraint="channels_pkey", set_=msg)
        )

    def update_last_update_channel(self, port, channel, timestamp):
        self.conn.execute(
            channels.update()
            .where((channels.c.port == port) & (channels.c.channel == channel))
            .values(last_update=timestamp)
        )

    def handle_set_relayer_tx_stat_days(self, port, channel, timestamp, address):
        relayer_tx_stat_day = {
            "date": timestamp,
            "port": port,
            "channel": channel,
            "address": address,
            "last_update_at": timestamp,
        }

        if (
            self.get_ibc_received_txs(
                relayer_tx_stat_day["date"],
                relayer_tx_stat_day["port"],
                relayer_tx_stat_day["channel"],
                relayer_tx_stat_day["address"],
            )
            is None
        ):
            relayer_tx_stat_day["ibc_received_txs"] = 1
            self.conn.execute(relayer_tx_stat_days.insert(), relayer_tx_stat_day)
        else:
            condition = True
            for col in relayer_tx_stat_days.primary_key.columns.values():
                condition = (col == relayer_tx_stat_day[col.name]) & condition
            self.conn.execute(
                relayer_tx_stat_days.update()
                .where(condition)
                .values(
                    ibc_received_txs=relayer_tx_stat_days.c.ibc_received_txs + 1,
                    last_update_at=timestamp,
                )
            )

    def handle_set_price_signal(self, msg):
        if msg["tx_hash"] is not None:
            msg["transaction_id"] = self.get_transaction_id(msg["tx_hash"])
        del msg["tx_hash"]
        msg["validator_id"] = self.get_validator_id(msg["validator"])
        del msg["validator"]
        msg["feeder_id"] = self.get_account_id(msg["feeder"])
        del msg["feeder"]
        self.conn.execute(
            insert(price_signals).values(**msg).on_conflict_do_update(constraint="price_signals_pkey", set_=msg)
        )

    def handle_set_validator_price(self, msg):
        msg["validator_id"] = self.get_validator_id(msg["validator"])
        del msg["validator"]
        self.conn.execute(
            insert(validator_prices)
            .values(**msg)
            .on_conflict_do_update(constraint="validator_prices_pkey", set_=msg)
        )

    def handle_remove_validator_prices(self, msg):
        self.conn.execute(
            validator_prices.delete().where(
                validator_prices.c.signal_id == msg["signal_id"]
            )
        )

    def handle_set_delegator_signal(self, msg):
        msg["account_id"] = self.get_account_id(msg["delegator"])
        del msg["delegator"]
        self.conn.execute(
            insert(delegator_signals)
            .values(**msg)
            .on_conflict_do_update(constraint="delegator_signals_pkey", set_=msg)
        )

    def handle_remove_delegator_signals(self, msg):
        self.conn.execute(
            delegator_signals.delete().where(
                delegator_signals.c.account_id == self.get_account_id(msg["delegator"])
            )
        )

    def handle_set_signal_total_power(self, msg):
        self.conn.execute(
            insert(signal_total_powers)
            .values(**msg)
            .on_conflict_do_update(constraint="signal_total_powers_pkey", set_=msg)
        )

    def handle_remove_signal_total_power(self, msg):
        self.conn.execute(
            signal_total_powers.delete().where(
                signal_total_powers.c.signal_id == msg["signal_id"]
            )
        )

    def handle_set_price(self, msg):
        self.conn.execute(
            insert(prices).values(**msg).on_conflict_do_update(constraint="prices_pkey", set_=msg)
        )
        self.conn.execute(
            prices.delete().where(prices.c.timestamp < msg["timestamp"] - PRICE_HISTORY_PERIOD)
        )

    def handle_remove_price(self, msg):
        self.conn.execute(prices.delete().where(prices.c.signal_id == msg["signal_id"]))

    def handle_set_reference_source_config(self, msg):
        self.conn.execute(
            insert(reference_source_configs).values(**msg).on_conflict_do_update(constraint="reference_source_configs_pkey", set_=msg)
        )
    
    def handle_set_feeder(self, msg):
        msg["operator_address"] = msg["validator"]
        del msg["validator"]
        msg["feeder_id"] = self.get_account_id(msg["feeder"])
        del msg["feeder"]
        self.conn.execute(
            insert(feeders)
            .values(msg)
            .on_conflict_do_nothing(constraint="feeders_pkey")
        )

    def handle_remove_feeder(self, msg):
        msg["operator_address"] = msg["validator"]
        del msg["validator"]
        msg["feeder_id"] = self.get_account_id(msg["feeder"])
        del msg["feeder"]
        condition = True
        for col in feeders.primary_key.columns.values():
            condition = (col == msg[col.name]) & condition
        self.conn.execute(feeders.delete().where(condition))
