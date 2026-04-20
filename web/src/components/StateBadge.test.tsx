import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { StatusBadge, WarrantyBadge } from "./StateBadge";

describe("StateBadge", () => {
  it("renders readable asset status labels", () => {
    render(<StatusBadge status="in_repair" />);
    expect(screen.getByText("In repair")).toBeInTheDocument();
  });

  it("renders warranty state labels", () => {
    render(<WarrantyBadge state="expiring_soon" />);
    expect(screen.getByText("Expiring soon")).toBeInTheDocument();
  });
});
