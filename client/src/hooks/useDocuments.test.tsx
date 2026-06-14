import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
// @ts-expect-error Vitest should load the TS source even when local JS mirror files exist.
import { useDocumentDetail } from "./useDocuments.ts";
import { fetchDocumentCourses, fetchDocumentDetail } from "../api/document";

vi.mock("../api/document", () => ({
  fetchDocumentCourses: vi.fn(),
  fetchDocumentDetail: vi.fn(),
}));

vi.mock("antd", async () => {
  const actual = await vi.importActual<typeof import("antd")>("antd");
  return {
    ...actual,
    message: {
      error: vi.fn(),
    },
  };
});

const mockedFetchDocumentCourses = vi.mocked(fetchDocumentCourses);
const mockedFetchDocumentDetail = vi.mocked(fetchDocumentDetail);

describe("useDocumentDetail", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedFetchDocumentCourses.mockResolvedValue({ data: [], total: 0 });
  });

  it("keeps the latest detail when an older request resolves later", async () => {
    let resolveOldRequest: (value: Awaited<ReturnType<typeof fetchDocumentDetail>>) => void = () => {};
    let resolveNewRequest: (value: Awaited<ReturnType<typeof fetchDocumentDetail>>) => void = () => {};

    mockedFetchDocumentDetail
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveOldRequest = resolve;
          }),
      )
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveNewRequest = resolve;
          }),
      );

    const { result, rerender } = renderHook(
      ({ courseId, docId }) => useDocumentDetail(courseId, docId),
      {
        initialProps: { courseId: "go", docId: "intro" },
      },
    );

    await waitFor(() => {
      expect(mockedFetchDocumentDetail).toHaveBeenCalledTimes(1);
    });

    rerender({ courseId: "go", docId: "advanced" });

    await waitFor(() => {
      expect(mockedFetchDocumentDetail).toHaveBeenCalledTimes(2);
    });

    await act(async () => {
      resolveNewRequest({
        data: {
          id: "advanced",
          title: "Advanced",
          order: 2,
          markdown: "new detail",
        },
      });
    });

    await waitFor(() => {
      expect(result.current.detail?.id).toBe("advanced");
      expect(result.current.loading).toBe(false);
    });

    await act(async () => {
      resolveOldRequest({
        data: {
          id: "intro",
          title: "Intro",
          order: 1,
          markdown: "old detail",
        },
      });
    });

    expect(result.current.detail?.id).toBe("advanced");
  });
});
