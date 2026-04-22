interface EditableTableDataProps {
    value: any;
    inputType: string;
    tdClass?: string;
    onSaveEdit: (value: any) => void;
}