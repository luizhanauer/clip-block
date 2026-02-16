export interface Clip {
    id: string;
    content: string;
    created_at: string; // Vem como string ISO do Go
    is_pinned: boolean;
}