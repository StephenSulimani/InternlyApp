import type { SavedSearch } from "../api/savedSearches";
import { boardStateMatchesSavedSearch } from "../lib/savedSearch";
import type { BoardFilters, BoardSort } from "../lib/boardFilters";
import styles from "./SavedSearchBar.module.css";

type Props = {
  searches: SavedSearch[];
  activeId: string | null;
  filters: BoardFilters;
  sort: BoardSort;
  disabled?: boolean;
  busy?: boolean;
  onSelect: (id: string | null) => void;
  onSave: (name: string) => void;
  onUpdate: () => void;
  onDelete: () => void;
};

export default function SavedSearchBar({
  searches,
  activeId,
  filters,
  sort,
  disabled = false,
  busy = false,
  onSelect,
  onSave,
  onUpdate,
  onDelete,
}: Props) {
  const active = activeId
    ? searches.find((search) => search.id === activeId)
    : undefined;
  const dirty =
    active != null && !boardStateMatchesSavedSearch(filters, sort, active);

  function handleSaveClick() {
    const name = window.prompt("Name this search");
    if (!name?.trim()) {
      return;
    }
    onSave(name.trim());
  }

  return (
    <div className={styles.bar}>
      <label className={styles.field}>
        <span>Saved search</span>
        <select
          value={activeId ?? ""}
          onChange={(event) => onSelect(event.target.value || null)}
          disabled={disabled || busy}
        >
          <option value="">Custom</option>
          {searches.map((search) => (
            <option key={search.id} value={search.id}>
              {search.name}
            </option>
          ))}
        </select>
      </label>

      <div className={styles.actions}>
        <button
          type="button"
          className={styles.action}
          onClick={handleSaveClick}
          disabled={disabled || busy}
        >
          Save current
        </button>

        {active && dirty ? (
          <button
            type="button"
            className={styles.action}
            onClick={onUpdate}
            disabled={disabled || busy}
          >
            Update “{active.name}”
          </button>
        ) : null}

        {active ? (
          <button
            type="button"
            className={`${styles.action} ${styles.danger}`}
            onClick={onDelete}
            disabled={disabled || busy}
          >
            Delete
          </button>
        ) : null}
      </div>
    </div>
  );
}
