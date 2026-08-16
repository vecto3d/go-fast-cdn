import { useCallback, useEffect, useRef, useState } from "react";

type Rect = { left: number; top: number; width: number; height: number };

// A drag under this many pixels is a click, not a marquee. Without it, every
// click-to-select would also run a (pointless) box selection.
const DRAG_THRESHOLD = 5;

type DragSelectOptions = {
  enabled: boolean;
  /** Called with the keys inside the box, continuously while dragging. */
  onSelect: (keys: string[], event: { shiftKey: boolean }) => void;
  /** Called once when a drag actually starts, to snapshot the selection. */
  onStart?: () => void;
};

/**
 * Marquee selection over any element carrying a data-selection-key attribute.
 * Returns the props to spread on the container and the box to render.
 */
export const useDragSelect = ({
  enabled,
  onSelect,
  onStart,
}: DragSelectOptions) => {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const origin = useRef<{ x: number; y: number } | null>(null);
  const dragging = useRef(false);
  const [box, setBox] = useState<Rect | null>(null);

  const keysWithin = useCallback((rect: Rect) => {
    const container = containerRef.current;
    if (!container) return [];

    const right = rect.left + rect.width;
    const bottom = rect.top + rect.height;

    return Array.from(
      container.querySelectorAll<HTMLElement>("[data-selection-key]")
    )
      .filter((element) => {
        const bounds = element.getBoundingClientRect();
        return (
          bounds.left < right &&
          bounds.right > rect.left &&
          bounds.top < bottom &&
          bounds.bottom > rect.top
        );
      })
      .map((element) => element.dataset.selectionKey as string);
  }, []);

  useEffect(() => {
    if (!enabled) return;

    const handleMove = (event: MouseEvent) => {
      if (!origin.current) return;

      const rect = {
        left: Math.min(origin.current.x, event.clientX),
        top: Math.min(origin.current.y, event.clientY),
        width: Math.abs(event.clientX - origin.current.x),
        height: Math.abs(event.clientY - origin.current.y),
      };

      if (
        !dragging.current &&
        rect.width < DRAG_THRESHOLD &&
        rect.height < DRAG_THRESHOLD
      ) {
        return;
      }

      if (!dragging.current) {
        dragging.current = true;
        onStart?.();
      }

      setBox(rect);
      onSelect(keysWithin(rect), { shiftKey: event.shiftKey });
    };

    const handleUp = () => {
      origin.current = null;
      setBox(null);
      // Cleared on the next tick so the click that ends the drag doesn't get
      // treated as a separate click-to-select on the card underneath.
      setTimeout(() => {
        dragging.current = false;
      }, 0);
    };

    window.addEventListener("mousemove", handleMove);
    window.addEventListener("mouseup", handleUp);

    return () => {
      window.removeEventListener("mousemove", handleMove);
      window.removeEventListener("mouseup", handleUp);
    };
  }, [enabled, keysWithin, onSelect, onStart]);

  const onMouseDown = (event: React.MouseEvent) => {
    // Left button only, and not when starting on a control inside a card.
    if (!enabled || event.button !== 0) return;
    if ((event.target as HTMLElement).closest("button, a, input, audio, video")) {
      return;
    }
    origin.current = { x: event.clientX, y: event.clientY };
  };

  return {
    containerRef,
    dragProps: { onMouseDown },
    /** True while a marquee drag is in progress. */
    isDragging: () => dragging.current,
    box,
  };
};
