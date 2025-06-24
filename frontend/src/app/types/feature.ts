import { User } from "./user";
import { Tag } from "./tag";

export interface Feature {
  id: number;
  project_id: number;
  parent_feature_id: number | null;
  title: string;
  description: string;
  status: 'todo' | 'in_progress' | 'done';
  priority: 'low' | 'medium' | 'high';
  assignee_id: number;
  created_at: string;
  updated_at: string;
  parent_feature?: Feature;
  assignee?: User;
  tags?: Tag[];
} 