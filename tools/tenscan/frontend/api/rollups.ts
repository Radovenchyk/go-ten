import { httpRequest } from ".";
import { apiRoutes } from "@/src/routes";
import { pathToUrl } from "@/src/routes/router";
import { ResponseDataInterface } from "@/src/types/interfaces";
import {
  Rollup,
  RollupsResponse,
} from "@/src/types/interfaces/RollupInterfaces";

export const fetchLatestRollups = async (
  payload?: Record<string, any>
): Promise<ResponseDataInterface<any>> => {
  return await httpRequest<ResponseDataInterface<any>>({
    method: "get",
    url: pathToUrl(apiRoutes.getLatestRollup),
    searchParams: payload,
  });
};

export const fetchRollups = async (): Promise<
  ResponseDataInterface<RollupsResponse>
> => {
  return await httpRequest<ResponseDataInterface<RollupsResponse>>({
    method: "get",
    url: pathToUrl(apiRoutes.getRollups),
  });
};

export const decryptEncryptedRollup = async ({
  StrData,
}: {
  StrData: string;
}): Promise<ResponseDataInterface<any>> => {
  return await httpRequest<ResponseDataInterface<any>>({
    method: "post",
    url: pathToUrl(apiRoutes.decryptEncryptedRollup),
    data: { StrData },
  });
};

export const fetchRollupByHash = async (
  hash: string
): Promise<ResponseDataInterface<Rollup>> => {
  return await httpRequest<ResponseDataInterface<Rollup>>({
    method: "get",
    url: pathToUrl(apiRoutes.getRollupByHash, { hash }),
  });
};

export const fetchRollupByBatchSequence = async (
  seq: string
): Promise<ResponseDataInterface<Rollup>> => {
  return await httpRequest<ResponseDataInterface<Rollup>>({
    method: "get",
    url: pathToUrl(apiRoutes.getRollupByBatchSequence, { seq }),
  });
};
