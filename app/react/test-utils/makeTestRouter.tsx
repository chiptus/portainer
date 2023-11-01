import { render, RenderOptions, RenderResult } from '@testing-library/react';
import {
  RawParams,
  ReactStateDeclaration,
  StateOrName,
  TransitionOptions,
  UIRouter,
  UIRouterReact,
  memoryLocationPlugin,
  servicesPlugin,
} from '@uirouter/react';
import { ReactElement } from 'react';

export function makeTestRouter(states: ReactStateDeclaration[]) {
  const router = new UIRouterReact();
  router.plugin(servicesPlugin);
  router.plugin(memoryLocationPlugin);
  router.locationConfig.html5Mode = () => true;
  states.forEach((state) => router.stateRegistry.register(state));

  function mountInRouter(
    ui: ReactElement,
    options?: Omit<RenderOptions, 'queries'>
  ): RenderResult {
    function WrapperComponent() {
      return <UIRouter router={router}>{ui}</UIRouter>;
    }

    return render(<WrapperComponent />, options);
  }

  function routerGo(
    to: StateOrName,
    params?: RawParams,
    options?: TransitionOptions
  ) {
    return router.stateService.go(to, params, options);
  }

  return { router, routerGo, mountInRouter };
}
