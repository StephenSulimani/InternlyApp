import type { BoardFilters, RecencyFilter } from "../lib/boardFilters";
import { EMPTY_BOARD_FILTERS, hasActiveFilters } from "../lib/boardFilters";
import ChipField from "./ChipField";
import styles from "./BoardToolbar.module.css";

type Props = {
  types: string[];
  locations: string[];
  sources: string[];
  filters: BoardFilters;
  onChange: (filters: BoardFilters) => void;
  disabled?: boolean;
};

const RECENCY_OPTIONS: { value: RecencyFilter; label: string }[] = [
  { value: "", label: "Any time" },
  { value: "24h", label: "Last 24 hours" },
  { value: "3d", label: "Last 3 days" },
  { value: "7d", label: "Last week" },
];

export default function BoardToolbar({
  types,
  locations,
  sources,
  filters,
  onChange,
  disabled = false,
}: Props) {
  const typeOptions = withSelected(types, filters.type);
  const sourceOptions = withSelected(sources, filters.source);
  const active = hasActiveFilters(filters);

  return (
    <div className={styles.toolbar}>
      <ChipField
        label="Search"
        value={filters.q}
        onChange={(q) => onChange({ ...filters, q })}
        placeholder="Role, company, or keyword"
        disabled={disabled}
      />

      <ChipField
        label="Location"
        value={filters.location}
        onChange={(location) => onChange({ ...filters, location })}
        placeholder="NYC, Manhattan, New York"
        suggestions={locations}
        disabled={disabled}
      />

      <div className={styles.filters}>
        <label className={styles.field}>
          <span>Type</span>
          <select
            value={filters.type}
            onChange={(event) =>
              onChange({ ...filters, type: event.target.value })
            }
            disabled={disabled}
          >
            <option value="">All types</option>
            {typeOptions.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </select>
        </label>

        <label className={styles.field}>
          <span>Source</span>
          <select
            value={filters.source}
            onChange={(event) =>
              onChange({ ...filters, source: event.target.value })
            }
            disabled={disabled}
          >
            <option value="">All sources</option>
            {sourceOptions.map((source) => (
              <option key={source} value={source}>
                {source}
              </option>
            ))}
          </select>
        </label>

        <label className={styles.field}>
          <span>Posted</span>
          <select
            value={filters.recency}
            onChange={(event) =>
              onChange({
                ...filters,
                recency: event.target.value as RecencyFilter,
              })
            }
            disabled={disabled}
          >
            {RECENCY_OPTIONS.map((option) => (
              <option key={option.value || "any"} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>

        <button
          type="button"
          className={
            filters.saved
              ? `${styles.savedToggle} ${styles.savedOn}`
              : styles.savedToggle
          }
          aria-pressed={filters.saved}
          onClick={() => onChange({ ...filters, saved: !filters.saved })}
          disabled={disabled}
        >
          Saved
        </button>

        {active && (
          <button
            type="button"
            className={styles.clear}
            onClick={() => onChange(EMPTY_BOARD_FILTERS)}
            disabled={disabled}
          >
            Clear
          </button>
        )}
      </div>
    </div>
  );
}

function withSelected(options: string[], selected: string): string[] {
  if (!selected || options.includes(selected)) {
    return options;
  }
  return [...options, selected].sort((a, b) => a.localeCompare(b));
}
