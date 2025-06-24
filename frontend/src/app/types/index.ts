export * from './user';
export * from './project';
export * from './feature';
export * from './subfeature';
export * from './tag';
export * from './task';

// Export Tag and Task interfaces
export interface Tag {
  tag_name: string;
  feature_id: number;
  created_by_user: number;
}

export interface Task {
  ID: number;
  task_type: string;
  task_name: string;
  description: string;
  feature_id: number;
}

// export * from './feature';
// export * from './user'; 