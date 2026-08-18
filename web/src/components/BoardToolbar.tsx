import {
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
} from "react";
import type { BoardFilters, RecencyFilter } from "../lib/boardFilters";
import {
  EMPTY_BOARD_FILTERS,
  addLocationToken,
  hasActiveFilters,
  joinLocationTokens,
  matchingLocationSuggestions,
  removeLocationToken,
  splitLocationTokens,
} from "../lib/boardFilters";
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
  const locationLabelId = useId();
  const locationWrapRef = useRef<HTMLDivElement>(null);
  const locationInputRef = useRef<HTMLInputElement>(null);
  const [locationDraft, setLocationDraft] = useState("");
  const [locationOpen, setLocationOpen] = useState(false);
  const [highlight, setHighlight] = useState(0);

  const selectedLocations = useMemo(
    () => splitLocationTokens(filters.location),
    [filters.location],
  );
  const typeOptions = withSelected(types, filters.type);
  const sourceOptions = withSelected(sources, filters.source);
  const active = hasActiveFilters(filters);
  const locationSuggestions = useMemo(
    () =>
      matchingLocationSuggestions(locations, selectedLocations, locationDraft),
    [locations, selectedLocations, locationDraft],
  );
  const showLocationSuggestions =
    !disabled && locationOpen && locationSuggestions.length > 0;

  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;
  const filtersRef = useRef(filters);
  filtersRef.current = filters;
  const draftRef = useRef(locationDraft);
  draftRef.current = locationDraft;
  const selectedRef = useRef(selectedLocations);
  selectedRef.current = selectedLocations;
  const commitDraftRef = useRef<(raw?: string) => boolean>(() => false);

  useEffect(() => {
    setHighlight(0);
  }, [locationSuggestions]);

  useEffect(() => {
    function onPointerDown(event: PointerEvent) {
      if (locationWrapRef.current?.contains(event.target as Node)) {
        return;
      }
      commitDraftRef.current(draftRef.current);
      setLocationOpen(false);
    }
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, []);

  function setLocations(tokens: string[]) {
    const next = joinLocationTokens(tokens);
    if (next === filtersRef.current.location) {
      return;
    }
    onChangeRef.current({ ...filtersRef.current, location: next });
  }

  function commitDraft(raw = locationDraft): boolean {
    const next = addLocationToken(selectedRef.current, raw);
    if (next.length === selectedRef.current.length) {
      if (raw.trim()) {
        setLocationDraft("");
      }
      return false;
    }
    setLocations(next);
    setLocationDraft("");
    return true;
  }
  commitDraftRef.current = commitDraft;

  function pickLocation(suggestion: string) {
    setLocations(addLocationToken(selectedRef.current, suggestion));
    setLocationDraft("");
    setLocationOpen(true);
    locationInputRef.current?.focus();
  }

  function removeLocation(token: string) {
    setLocations(removeLocationToken(selectedRef.current, token));
    locationInputRef.current?.focus();
  }

  function onLocationKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === "Escape") {
      setLocationOpen(false);
      return;
    }

    if (event.key === "Backspace" && locationDraft === "") {
      const last = selectedLocations[selectedLocations.length - 1];
      if (last) {
        event.preventDefault();
        removeLocation(last);
      }
      return;
    }

    if (event.key === ",") {
      event.preventDefault();
      if (commitDraft()) {
        setLocationOpen(true);
      }
      return;
    }

    if (event.key === "ArrowDown") {
      event.preventDefault();
      setLocationOpen(true);
      setHighlight((index) =>
        showLocationSuggestions
          ? (index + 1) % locationSuggestions.length
          : 0,
      );
      return;
    }

    if (event.key === "ArrowUp" && showLocationSuggestions) {
      event.preventDefault();
      setHighlight(
        (index) =>
          (index - 1 + locationSuggestions.length) % locationSuggestions.length,
      );
      return;
    }

    if (event.key !== "Enter") {
      return;
    }

    event.preventDefault();
    if (showLocationSuggestions) {
      pickLocation(locationSuggestions[highlight] ?? locationSuggestions[0]);
      return;
    }
    commitDraft();
  }

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

      <div className={styles.search} ref={locationWrapRef}>
        <span className={styles.searchLabel} id={locationLabelId}>
          Location
        </span>
        <div className={styles.chipWrap}>
          <div
            className={
              disabled
                ? `${styles.chipField} ${styles.chipFieldDisabled}`
                : styles.chipField
            }
            onClick={() => locationInputRef.current?.focus()}
          >
            {selectedLocations.map((location) => (
              <span key={location} className={styles.chip}>
                <span className={styles.chipLabel}>{location}</span>
                <button
                  type="button"
                  className={styles.chipRemove}
                  aria-label={`Remove ${location}`}
                  disabled={disabled}
                  onClick={(event) => {
                    event.stopPropagation();
                    removeLocation(location);
                  }}
                >
                  ×
                </button>
              </span>
            ))}
            <input
              ref={locationInputRef}
              className={styles.chipInput}
              type="text"
              value={locationDraft}
              onChange={(event) => {
                setLocationDraft(event.target.value);
                setLocationOpen(true);
              }}
              onFocus={() => setLocationOpen(true)}
              onKeyDown={onLocationKeyDown}
              placeholder={
                selectedLocations.length === 0
                  ? "NYC, Manhattan, New York"
                  : "Add another"
              }
              autoComplete="off"
              role="combobox"
              aria-labelledby={locationLabelId}
              aria-autocomplete="list"
              aria-expanded={showLocationSuggestions}
              aria-controls={locationListId}
              aria-activedescendant={
                showLocationSuggestions
                  ? `${locationListId}-${highlight}`
                  : undefined
              }
              disabled={disabled}
            />
          </div>
          {showLocationSuggestions && (
            <ul
              className={styles.suggestions}
              id={locationListId}
              role="listbox"
            >
              {locationSuggestions.map((location, index) => (
                <li key={location} role="presentation">
                  <button
                    type="button"
                    id={`${locationListId}-${index}`}
                    role="option"
                    aria-selected={index === highlight}
                    className={
                      index === highlight
                        ? `${styles.suggestion} ${styles.suggestionActive}`
                        : styles.suggestion
                    }
                    onMouseEnter={() => setHighlight(index)}
                    onMouseDown={(event) => event.preventDefault()}
                    onClick={() => pickLocation(location)}
                  >
                    {location}
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

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
            onClick={() => {
              setLocationDraft("");
              onChange(EMPTY_BOARD_FILTERS);
            }}
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
