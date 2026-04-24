interface CueTableProps {
    CueListId:  string;
    onEdit: (cueListId: string, cueId: string) => void;
}