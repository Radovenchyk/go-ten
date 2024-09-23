import TruncatedAddress from "@repo/ui/common/truncated-address";
import { formatTimeAgo } from "@repo/ui/lib/utils";
import { Batch } from "@/src/types/interfaces/BatchInterfaces";
import { Avatar, AvatarFallback } from "@repo/ui/shared/avatar";
import Link from "next/link";

export function RecentBatches({ batches }: { batches: any }) {
  return (
    <div className="space-y-8">
      {batches?.result?.BatchesData.map((batch: Batch, i: number) => (
        <div className="flex items-center" key={i}>
          <Avatar className="h-9 w-9">
            <AvatarFallback>BN</AvatarFallback>
          </Avatar>
          <div className="ml-4 space-y-1">
            <p className="text-sm font-medium leading-none">
              <Link
                href={`/batch/height/${batch?.height}`}
                className="text-primary"
              >
                #{Number(batch?.height)}
              </Link>
            </p>
            <p className="text-sm text-muted-foreground word-break-all">
              {formatTimeAgo(batch?.header?.timestamp)}
            </p>
          </div>
          <div className="ml-auto font-medium min-w-[140px]">
            <Link href={`/batch/${batch?.hash}`} className="text-primary">
              <TruncatedAddress address={batch?.hash} />
            </Link>
          </div>
        </div>
      ))}
    </div>
  );
}
