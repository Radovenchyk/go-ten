import TruncatedAddress from "@repo/ui/common/truncated-address";
import KeyValueItem, { KeyValueList } from "@repo/ui/shared/key-value";
import { formatTimeAgo } from "@repo/ui/lib/utils";
import { Rollup } from "@/src/types/interfaces/RollupInterfaces";
import Link from "next/link";

export function RollupDetailsComponent({
  rollupDetails,
}: {
  rollupDetails: Rollup;
}) {
  return (
    <div className="space-y-8">
      <KeyValueList>
        <KeyValueItem label="ID" value={"#" + Number(rollupDetails?.ID)} />
        <KeyValueItem
          label="Timestamp"
          value={formatTimeAgo(rollupDetails?.Timestamp)}
        />
        <KeyValueItem
          label="Full Hash"
          value={
            <Link
              href={`/rollup/${rollupDetails?.Hash}`}
              className="text-primary"
            >
              <TruncatedAddress address={rollupDetails?.Hash} />
            </Link>
          }
        />
        <KeyValueItem
          label="Rollup Header Hash"
          value={
            <Link
              href={`/rollup/${rollupDetails?.Header?.hash}`}
              className="text-primary"
            >
              <TruncatedAddress address={rollupDetails?.Header?.hash} />
            </Link>
          }
        />
        <KeyValueItem
          label="L1 Hash"
          value={<TruncatedAddress address={rollupDetails?.L1Hash} />}
        />
        <KeyValueItem
          label="First Sequencer"
          value={
            <Link
              href={`/rollup/batch/sequence/${rollupDetails?.FirstSeq}`}
              className="text-primary"
            >
              {"#" + rollupDetails?.FirstSeq}
            </Link>
          }
        />
        <KeyValueItem
          label="Last Sequencer"
          value={
            <Link
              href={`/rollup/batch/sequence/${rollupDetails?.LastSeq}`}
              className="text-primary"
            >
              {"#" + rollupDetails?.LastSeq}
            </Link>
          }
        />
        <KeyValueItem
          label="Compression L1 Head"
          value={
            <TruncatedAddress
              address={rollupDetails?.Header?.CompressionL1Head}
            />
          }
        />
        <KeyValueItem
          label="Payload Hash"
          value={
            <TruncatedAddress address={rollupDetails?.Header?.PayloadHash} />
          }
        />
        <KeyValueItem
          label="Signature"
          value={
            <TruncatedAddress address={rollupDetails?.Header?.Signature} />
          }
        />
        <KeyValueItem
          label="Last Batch Sequence No"
          value={
            <Link
              href={`/rollup/batch/sequence/${rollupDetails?.Header?.LastBatchSeqNo}`}
              className="text-primary"
            >
              {"#" + rollupDetails?.Header?.LastBatchSeqNo}
            </Link>
          }
        />
        <KeyValueItem
          label="Cross Chain Messages"
          value={
            rollupDetails?.Header?.crossChainMessages.length > 0
              ? rollupDetails?.Header?.crossChainMessages?.map((msg, index) => (
                  <div key={index} className="space-y-4">
                    <KeyValueList>
                      <KeyValueItem label="Sender" value={msg.Sender} />
                      <KeyValueItem
                        label="Sequence"
                        value={
                          <Link
                            href={`/rollup/batch/sequence/${msg.Sequence}`}
                            className="text-primary"
                          >
                            {"#" + msg.Sequence}
                          </Link>
                        }
                      />
                      <KeyValueItem label="Nonce" value={msg.Nonce} />
                      <KeyValueItem label="Topic" value={msg.Topic} />
                      <KeyValueItem
                        label="Payload"
                        value={msg.Payload.map((payload, index) => (
                          <div key={index}>{payload}</div>
                        ))}
                      />
                      <KeyValueItem
                        label="Consistency Level"
                        value={msg.ConsistencyLevel}
                      />
                    </KeyValueList>
                  </div>
                ))
              : "No cross chain messages found."
          }
          isLastItem
        />
      </KeyValueList>
    </div>
  );
}
