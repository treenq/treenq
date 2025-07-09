type FetchFn = typeof fetch;

type RequestOptions = Omit<RequestInit, "method"> & {
  query?: Record<string, string | number | boolean>;
  headers?: Record<string, string>;
};

export type Failure = {
  error: ApiErrorPayload;
};

export type Success<T = void> = {
  data: T;
};

export type Result<T = void> = Failure | Success<T>;

export type ApiErrorPayload = {
  code: string;
  message: string;
  meta: Record<string, string>;
};
export type TestTypeNoJsonTags = {
  Value: string;
};

export type TestTypeNestedTypes = {
  data: TestStruct;
  chunk: number[];
  slice: HighElem[];
  map: Record<number, HighMapElem>;
  nextP: HighPointer | undefined;
};

export type TestStruct = {
  row: number;
  line: string;
  next: TestNextLevelStruct;
  slice: TestNextLevelElem[];
  map: Record<number, MapValue>;
  nextP: TestNextLevelStructP | undefined;
};

export type TestNextLevelStruct = {
  extra: string;
};

export type TestNextLevelElem = {
  int: number;
};

export type MapValue = {
  Value: string;
};

export type TestNextLevelStructP = {
  extra: string;
};

export type HighElem = {
  int: number;
};

export type HighMapElem = {
  Value: string;
};

export type HighPointer = {
  extra: string;
};

export type GetQuery = {
  Value: string;
  Field: number;
};

export type GetResp = {
  Getting: number;
};

export type TimeTestRequest = {
  createdAt: string;
  name: string;
};

export type TimeTestResponse = {
  processedAt: string;
  id: string;
};

class Client {
  constructor(
    private baseUrl: string,
    private fetchFn: FetchFn = window.fetch.bind(window),
  ) {
    if (!baseUrl.endsWith("/")) {
      baseUrl = baseUrl + "/";
    }
    this.baseUrl = baseUrl;
    this.fetchFn = fetchFn;
  }

  private buildUrl(
    path: string,
    query?: Record<string, string | number | boolean>,
  ): string {
    if (path.startsWith("/")) {
      path = path.slice(1);
    }
    const url = new URL(path, this.baseUrl);
    if (query) {
      for (const [key, val] of Object.entries(query)) {
        url.searchParams.set(key, String(val));
      }
    }
    return url.toString();
  }

  private async request<T>(
    method: string,
    path: string,
    opts: RequestOptions = {},
  ): Promise<Result<T>> {
    const url = this.buildUrl(path, opts.query);
    const res = await this.fetchFn(url, {
      method,
      credentials: "include",
      ...opts,
      headers: {
        "Content-Type": "application/json",
        ...opts.headers,
      },
    });

    if (!res.ok) {
      if (res.status >= 500) {
        const errText = await res.text();
        throw Error("http error: " + errText);
      }
      const jsonErr = await res.json();
      return { error: jsonErr as ApiErrorPayload };
    }

    const response = await res.text();
    if (response) {
      const resp = JSON.parse(response);
      return { data: resp as T };
    }
    return { data: {} as T };
  }

  private async post<T>(
    path: string,
    body?: unknown,
    opts?: RequestOptions,
  ): Promise<Result<T>> {
    return await this.request("POST", path, {
      ...opts,
      body: JSON.stringify(body),
    });
  }

  private async get<T>(
    path: string,
    opts?: RequestOptions,
  ): Promise<Result<T>> {
    return await this.request("GET", path, opts);
  }
  async Test1(req: TestTypeNoJsonTags): Promise<Result<TestTypeNoJsonTags>> {
    return await this.post("test1", req);
  }

  async Test2(req: TestTypeNestedTypes): Promise<Result<TestTypeNestedTypes>> {
    return await this.post("test2", req);
  }

  async TestEmpty(): Promise<Result<void>> {
    return await this.post("testEmpty");
  }

  async TestGet(req: GetQuery): Promise<Result<GetResp>> {
    const query: Record<string, string | number | boolean> = {};
    query["value"] = req.Value;
    query["field"] = req.Field;
    return await this.get("testGet", { query });
  }

  async TestTime(req: TimeTestRequest): Promise<Result<TimeTestResponse>> {
    return await this.post("testTime", req);
  }
}
