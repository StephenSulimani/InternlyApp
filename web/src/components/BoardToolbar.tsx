import { useId, useMemo } from "react";
import type { BoardFilters, RecencyFilter } from "../lib/boardFilters";
import { EMPTY_BOARD_FILTERS, hasActiveFilters } from "../lib/boardFilters";
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
  const locationListId = useId();
  const typeOptions = withSelected(types, filters.type);
  const sourceOptions = withSelected(sources, filters.source);
  const active = hasActiveFilters(filters);
  const locationSuggestions = useMemo(
    () => matchingLocations(locations, filters.location),
    [locations, filters.location],
  );

  return (
    <div className={styles.toolbar}>
      <label className={styles.search}>
        <span className={styles.searchLabel}>Search</span>
        <input
          type="search"
          value={filters.q}
          onChange={(event) => onChange({ ...filters, q: event.target.value })}
          placeholder="Role, company, or keyword"
          autoComplete="off"
          disabled={disabled}
        />
      </label>

      <label className={styles.search}>
        <span className={styles.searchLabel}>Location</span>
        <input
          type="search"
          value={filters.location}
          onChange={(event) =>
            onChange({ ...filters, location: event.target.value })
          }
          placeholder="City, remote, or office"
          autoComplete="off"
          list={locationListId}
          disabled={disabled}
        />
        <datalist id={locationListId}>
          {locationSuggestions.map((location) => (
            <option key={location} value={location} />
          ))}
        </datalist>
      </label>

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

function matchingLocations(locations: string[], query: string): string[] {
  const needle = query.trim().toLowerCase();
  const matches = needle
    ? locations.filter((location) => location.toLowerCase().includes(needle))
    : locations;
  return matches.slice(0, 20);
}

function withSelected(options: string[], selected: string): string[] {
  if (!selected || options.includes(selected)) {
    return options;
  }
  return [...options, selected].sort((a, b) => a.localeCompare(b));
}
