import {
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
  type KeyboardEvent,
} from "react";
import {
  addLocationToken,
  joinLocationTokens,
  matchingLocationSuggestions,
  removeLocationToken,
  splitLocationTokens,
} from "../lib/boardFilters";
import styles from "./ChipField.module.css";

type Props = {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder: string;
  disabled?: boolean;
  suggestions?: string[];
};

export default function ChipField({
  label,
  value,
  onChange,
  placeholder,
  disabled = false,
  suggestions,
}: Props) {
  const listId = useId();
  const labelId = useId();
  const wrapRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const [draft, setDraft] = useState("");
  const [open, setOpen] = useState(false);
  const [highlight, setHighlight] = useState(0);

  const tokens = useMemo(() => splitLocationTokens(value), [value]);
  const matches = useMemo(
    () =>
      suggestions
        ? matchingLocationSuggestions(suggestions, tokens, draft)
        : [],
    [suggestions, tokens, draft],
  );
  const showSuggestions = !disabled && Boolean(suggestions) && open && matches.length > 0;

  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;
  const tokensRef = useRef(tokens);
  tokensRef.current = tokens;
  const draftRef = useRef(draft);
  draftRef.current = draft;
  const commitDraftRef = useRef<(raw?: string) => boolean>(() => false);

  useEffect(() => {
    setHighlight(0);
  }, [matches]);

  useEffect(() => {
    if (value === "") {
      setDraft("");
    }
  }, [value]);

  useEffect(() => {
    function onPointerDown(event: PointerEvent) {
      if (wrapRef.current?.contains(event.target as Node)) {
        return;
      }
      commitDraftRef.current(draftRef.current);
      setOpen(false);
    }
    document.addEventListener("pointerdown", onPointerDown);
    return () => document.removeEventListener("pointerdown", onPointerDown);
  }, []);

  function setTokens(nextTokens: string[]) {
    const next = joinLocationTokens(nextTokens);
    if (next === value) {
      return;
    }
    onChangeRef.current(next);
  }

  function commitDraft(raw = draft): boolean {
    const next = addLocationToken(tokensRef.current, raw);
    if (next.length === tokensRef.current.length) {
      if (raw.trim()) {
        setDraft("");
      }
      return false;
    }
    setTokens(next);
    setDraft("");
    return true;
  }
  commitDraftRef.current = commitDraft;

  function pickSuggestion(suggestion: string) {
    setTokens(addLocationToken(tokensRef.current, suggestion));
    setDraft("");
    setOpen(true);
    inputRef.current?.focus();
  }

  function removeToken(token: string) {
    setTokens(removeLocationToken(tokensRef.current, token));
    inputRef.current?.focus();
  }

  function onKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === "Escape") {
      setOpen(false);
      return;
    }

    if (event.key === "Backspace" && draft === "") {
      const last = tokens[tokens.length - 1];
      if (last) {
        event.preventDefault();
        removeToken(last);
      }
      return;
    }

    if (event.key === ",") {
      event.preventDefault();
      if (commitDraft()) {
        setOpen(true);
      }
      return;
    }

    if (event.key === "ArrowDown" && matches.length > 0) {
      event.preventDefault();
      setOpen(true);
      setHighlight((index) =>
        showSuggestions ? (index + 1) % matches.length : 0,
      );
      return;
    }

    if (event.key === "ArrowUp" && showSuggestions) {
      event.preventDefault();
      setHighlight((index) => (index - 1 + matches.length) % matches.length);
      return;
    }

    if (event.key !== "Enter") {
      return;
    }

    event.preventDefault();
    if (showSuggestions) {
      pickSuggestion(matches[highlight] ?? matches[0]);
      return;
    }
    commitDraft();
  }

  return (
    <div className={styles.field} ref={wrapRef}>
      <span className={styles.label} id={labelId}>
        {label}
      </span>
      <div className={styles.chipWrap}>
        <div
          className={
            disabled ? `${styles.chipField} ${styles.chipFieldDisabled}` : styles.chipField
          }
          onClick={() => inputRef.current?.focus()}
        >
          {tokens.map((token) => (
            <span key={token} className={styles.chip}>
              <span className={styles.chipLabel}>{token}</span>
              <button
                type="button"
                className={styles.chipRemove}
                aria-label={`Remove ${token}`}
                disabled={disabled}
                onClick={(event) => {
                  event.stopPropagation();
                  removeToken(token);
                }}
              >
                ×
              </button>
            </span>
          ))}
          <input
            ref={inputRef}
            className={styles.chipInput}
            type="text"
            value={draft}
            onChange={(event) => {
              setDraft(event.target.value);
              setOpen(true);
            }}
            onFocus={() => setOpen(true)}
            onKeyDown={onKeyDown}
            placeholder={tokens.length === 0 ? placeholder : "Add another"}
            autoComplete="off"
            role={suggestions ? "combobox" : undefined}
            aria-labelledby={labelId}
            aria-autocomplete={suggestions ? "list" : undefined}
            aria-expanded={suggestions ? showSuggestions : undefined}
            aria-controls={suggestions ? listId : undefined}
            aria-activedescendant={
              showSuggestions ? `${listId}-${highlight}` : undefined
            }
            disabled={disabled}
          />
        </div>
        {showSuggestions && (
          <ul className={styles.suggestions} id={listId} role="listbox">
            {matches.map((match, index) => (
              <li key={match} role="presentation">
                <button
                  type="button"
                  id={`${listId}-${index}`}
                  role="option"
                  aria-selected={index === highlight}
                  className={
                    index === highlight
                      ? `${styles.suggestion} ${styles.suggestionActive}`
                      : styles.suggestion
                  }
                  onMouseEnter={() => setHighlight(index)}
                  onMouseDown={(event) => event.preventDefault()}
                  onClick={() => pickSuggestion(match)}
                >
                  {match}
                </button>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
