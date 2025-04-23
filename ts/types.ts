export type Component<T extends Element> = {
  id: string;
  mount: (el: T) => (() => void) | undefined;
};
