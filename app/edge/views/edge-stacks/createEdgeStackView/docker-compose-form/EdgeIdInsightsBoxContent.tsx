export function EdgeIdInsightsBoxContent() {
  return (
    <div>
      You can use it as environment variable with{' '}
      <b>image: myapp:$&#123;PORTAINER_EDGE_ID&#125;</b> or use it with relative
      path for volumes{' '}
      <b>- ./config/$&#123;PORTAINER_EDGE_ID&#125;:/myapp/config</b>.
      <br />
      More documentation can be found{' '}
      <a href="https://docs.portainer.io/user/edge/configurations">here</a>.
    </div>
  );
}
