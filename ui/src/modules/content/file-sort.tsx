import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Sort,
  SortKey,
  SORT_DIRECTION_LABELS,
  SORT_LABELS,
} from "@/lib/file-sort";
import { ArrowDownAZ, ArrowUpAZ } from "lucide-react";

type FileSortProps = {
  sort: Sort;
  onChange: (sort: Sort) => void;
};

const FileSort: React.FC<FileSortProps> = ({ sort, onChange }) => (
  <div className="flex items-center gap-2">
    <Select
      value={sort.key}
      onValueChange={(key) => onChange({ ...sort, key: key as SortKey })}
    >
      <SelectTrigger className="w-[130px]" aria-label="Sort by">
        <SelectValue placeholder="Sort by" />
      </SelectTrigger>
      <SelectContent>
        {(Object.keys(SORT_LABELS) as SortKey[]).map((key) => (
          <SelectItem key={key} value={key}>
            {SORT_LABELS[key]}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
    <Button
      variant="outline"
      onClick={() =>
        onChange({
          ...sort,
          direction: sort.direction === "asc" ? "desc" : "asc",
        })
      }
      title={`Sorted ${SORT_DIRECTION_LABELS[sort.key][sort.direction]}`}
      aria-label={`Sorted ${SORT_DIRECTION_LABELS[sort.key][sort.direction]}. Click to reverse.`}
    >
      {sort.direction === "asc" ? <ArrowDownAZ /> : <ArrowUpAZ />}
      {SORT_DIRECTION_LABELS[sort.key][sort.direction]}
    </Button>
  </div>
);

export default FileSort;
