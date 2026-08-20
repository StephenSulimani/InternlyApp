import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import Navbar from "../components/Navbar";
import StatusBar from "../components/StatusBar";
import Footer from "../components/Footer";
import PageShell from "../components/PageShell";
import SavedSearchBar from "../components/SavedSearchBar";
import BoardToolbar from "../components/BoardToolbar";
import ListingCard from "../components/ListingCard";
import JobDetailModal from "../components/JobDetailModal";
import { useBoardListings } from "../hooks/useBoardListings";
import { usePersistedBoardSearch } from "../hooks/usePersistedBoardSearch";
import {
  useSavedSearchMutations,
  useSavedSearches,
} from "../hooks/useSavedSearches";
import { useToggleSavedJob } from "../hooks/useToggleSavedJob";
import {
  boardStateToSavedSearchInput,
  savedSearchToBoardState,
} from "../lib/savedSearch";
import {
  EMPTY_BOARD_FILTERS,
  hasActiveFilters,
  type BoardFilters,
  type BoardSort,
  type BoardSortField,
} from "../lib/boardFilters";
import type { Listing } from "../lib/listings";
import { useAuth } from "../providers/AuthProvider";
import styles from "./BoardPage.module.css";

export default function BoardPage() {
  const { user, isAuthenticated } = useAuth();
  const {
    filters,
    setFilters,
    sort,
    setSort,
    activeSavedSearchId,
    setActiveSavedSearchId,
  } = usePersistedBoardSearch(user?.id);
  const [offset, setOffset] = useState(0);
  const [selected, setSelected] = useState<Listing | null>(null);
  const savedSearchesQuery = useSavedSearches(isAuthenticated);
  const savedSearchMutations = useSavedSearchMutations();
  const savedSearches = savedSearchesQuery.data ?? [];
  const savedSearchBusy =
    savedSearchMutations.create.isPending ||
    savedSearchMutations.update.isPending ||
    savedSearchMutations.remove.isPending;

  useEffect(() => {
    setOffset(0);
    setSelected(null);
  }, [user?.id]);

  const {
    listings,
    total,
    limit,
    isPending,
    isFetching,
    isError,
    facets,
  } = useBoardListings(filters, sort, offset, isAuthenticated);
  const toggleSave = useToggleSavedJob();

  useEffect(() => {
    if (!selected) {
      return;
    }
    const fresh = listings.find((listing) => listing.id === selected.id);
    if (fresh && fresh !== selected) {
      setSelected(fresh);
    }
  }, [listings, selected]);

  const closeModal = useCallback(() => setSelected(null), []);

  function handleToggleSave(listing: Listing) {
    toggleSave.mutate(
      { id: listing.id, saved: Boolean(listing.saved) },
      {
        onSuccess: (nextSaved) => {
          setSelected((current) => {
            if (!current || current.id !== listing.id) {
              return current;
            }
            if (filters.saved && !nextSaved) {
              return null;
            }
            return { ...current, saved: nextSaved };
          });
        },
      },
    );
  }

  function handleFiltersChange(next: BoardFilters) {
    setActiveSavedSearchId(null);
    setFilters(next);
    setOffset(0);
    setSelected(null);
  }

  function handleSort(field: BoardSortField) {
    setActiveSavedSearchId(null);
    setSort((current) => {
      if (current.field === field) {
        return { field, dir: current.dir === "asc" ? "desc" : "asc" };
      }
      return { field, dir: field === "posted" ? "desc" : "asc" };
    });
    setOffset(0);
    setSelected(null);
  }

  function handleSelectSavedSearch(id: string | null) {
    if (!id) {
      setActiveSavedSearchId(null);
      return;
    }
    const search = savedSearches.find((item) => item.id === id);
    if (!search) {
      return;
    }
    const next = savedSearchToBoardState(search);
    setActiveSavedSearchId(id);
    setFilters(next.filters);
    setSort(next.sort);
    setOffset(0);
    setSelected(null);
  }

  function handleSaveSearch(name: string) {
    savedSearchMutations.create.mutate(
      boardStateToSavedSearchInput(name, filters, sort),
      {
        onSuccess: (created) => {
          setActiveSavedSearchId(created.id);
        },
      },
    );
  }

  function handleUpdateSavedSearch() {
    if (!activeSavedSearchId) {
      return;
    }
    const active = savedSearches.find(
      (search) => search.id === activeSavedSearchId,
    );
    if (!active) {
      return;
    }
    savedSearchMutations.update.mutate({
      id: activeSavedSearchId,
      input: boardStateToSavedSearchInput(active.name, filters, sort),
    });
  }

  function handleDeleteSavedSearch() {
    if (!activeSavedSearchId) {
      return;
    }
    const active = savedSearches.find(
      (search) => search.id === activeSavedSearchId,
    );
    if (
      !window.confirm(
        active ? `Delete saved search “${active.name}”?` : "Delete saved search?",
      )
    ) {
      return;
    }
    savedSearchMutations.remove.mutate(activeSavedSearchId, {
      onSuccess: () => setActiveSavedSearchId(null),
    });
  }

  const from = total === 0 ? 0 : offset + 1;
  const to = offset + listings.length;
  const canPrev = offset > 0;
  const canNext = offset + listings.length < total;
  const filtersActive = hasActiveFilters(filters);

  let countLabel = "Sign in to browse";
  if (isAuthenticated) {
    if (isPending && listings.length === 0) {
      countLabel = "Loading…";
    } else if (total === 0) {
      countLabel = "0 roles";
    } else {
      countLabel = `${from}–${to} of ${total} roles`;
    }
    if (isFetching && listings.length > 0) {
      countLabel = `${countLabel} · updating`;
    }
  }

  return (
    <PageShell>
      <Navbar />
      <StatusBar />
      <main className={styles.main}>
        <section className={styles.board}>
          <header className={styles.header}>
            <div>
              <span className={styles.label}>§ The board</span>
              <h1 className={styles.title}>The full index</h1>
              <p className={styles.lede}>
                Search and filter internships, new-grad roles, and early-career
                openings. Click a listing for details and apply on the original
                posting.
              </p>
            </div>
            <p className={styles.count}>{countLabel}</p>
          </header>

          <SavedSearchBar
            searches={savedSearches}
            activeId={activeSavedSearchId}
            filters={filters}
            sort={sort}
            disabled={!isAuthenticated}
            busy={savedSearchBusy}
            onSelect={handleSelectSavedSearch}
            onSave={handleSaveSearch}
            onUpdate={handleUpdateSavedSearch}
            onDelete={handleDeleteSavedSearch}
          />

          <BoardToolbar
            types={facets.types}
            locations={facets.locations}
            filters={filters}
            onChange={handleFiltersChange}
            disabled={!isAuthenticated}
          />

          <div className={styles.catalog}>
            <div className={styles.catalogHead}>
              <span className={styles.headIndex} aria-hidden="true">
                #
              </span>
              <SortButton
                field="company"
                label="Company / Role"
                sort={sort}
                onSort={handleSort}
                disabled={!isAuthenticated}
              />
              <SortButton
                field="location"
                label="Location"
                sort={sort}
                onSort={handleSort}
                disabled={!isAuthenticated}
              />
              <SortButton
                field="type"
                label="Type"
                sort={sort}
                onSort={handleSort}
                disabled={!isAuthenticated}
              />
              <SortButton
                field="posted"
                label="Posted"
                sort={sort}
                onSort={handleSort}
                disabled={!isAuthenticated}
              />
              <span className={styles.headHint} aria-hidden="true" />
            </div>

            {!isAuthenticated && (
              <p className={styles.empty}>
                Sign in to search and filter the full board.{" "}
                <Link to="/login" className={styles.reset}>
                  Sign in
                </Link>
              </p>
            )}

            {isAuthenticated && isPending && listings.length === 0 && (
              <p className={styles.empty}>Loading listings…</p>
            )}

            {isAuthenticated && isError && listings.length === 0 && (
              <p className={styles.empty}>
                Couldn’t load the board. Try again in a moment.
              </p>
            )}

            {isAuthenticated &&
              !isPending &&
              !isError &&
              listings.length === 0 && (
                <p className={styles.empty}>
                  {filters.saved &&
                  !filters.q.trim() &&
                  !filters.type &&
                  !filters.location &&
                  !filters.recency ? (
                    "No saved roles yet. Heart a listing to keep it here."
                  ) : filtersActive ? (
                    <>
                      No roles match those filters. Try a broader search, or{" "}
                      <button
                        type="button"
                        className={styles.reset}
                        onClick={() => handleFiltersChange(EMPTY_BOARD_FILTERS)}
                      >
                        clear filters
                      </button>
                      .
                    </>
                  ) : (
                    "No listings yet — check back soon."
                  )}
                </p>
              )}

            {isAuthenticated &&
              listings.map((listing, index) => (
                <ListingCard
                  key={listing.id}
                  listing={listing}
                  index={offset + index + 1}
                  selected={selected?.id === listing.id}
                  onSelect={setSelected}
                  onToggleSave={handleToggleSave}
                  saveDisabled={toggleSave.isPending}
                />
              ))}
          </div>

          {isAuthenticated && total > 0 && (
            <nav className={styles.pager} aria-label="Board pagination">
              <button
                type="button"
                className={styles.pageBtn}
                onClick={() => {
                  setSelected(null);
                  setOffset(Math.max(0, offset - limit));
                }}
                disabled={!canPrev || isFetching}
              >
                Previous
              </button>
              <span className={styles.pageStatus}>
                {from}–{to} of {total}
              </span>
              <button
                type="button"
                className={styles.pageBtn}
                onClick={() => {
                  setSelected(null);
                  setOffset(offset + limit);
                }}
                disabled={!canNext || isFetching}
              >
                Next
              </button>
            </nav>
          )}
        </section>
      </main>
      <Footer />
      <JobDetailModal
        listing={selected}
        onClose={closeModal}
        onToggleSave={handleToggleSave}
        saveDisabled={toggleSave.isPending}
      />
    </PageShell>
  );
}

function SortButton({
  field,
  label,
  sort,
  onSort,
  disabled,
}: {
  field: BoardSortField;
  label: string;
  sort: BoardSort;
  onSort: (field: BoardSortField) => void;
  disabled: boolean;
}) {
  const active = sort.field === field;
  const arrow = active ? (sort.dir === "asc" ? "↑" : "↓") : "";

  return (
    <button
      type="button"
      className={`${styles.sortBtn} ${active ? styles.sortActive : ""}`}
      onClick={() => onSort(field)}
      disabled={disabled}
      aria-pressed={active}
      aria-label={`Sort by ${label}${active ? `, ${sort.dir}ending` : ""}`}
    >
      {label}
      {arrow ? <span aria-hidden="true">{arrow}</span> : null}
    </button>
  );
}
