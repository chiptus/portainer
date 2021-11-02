import * as moment from 'moment-timezone';

export default class TimeWindowDisplayController {
  /* @ngInject */
  constructor($async, EndpointProvider, EndpointService) {
    this.$async = $async;
    this.EndpointProvider = EndpointProvider;
    this.EndpointService = EndpointService;
  }

  utcToTime(utcTime) {
    const startTime = moment.tz(utcTime.StartTime, 'HH:mm', 'GMT').tz(this.state.timezone).format('HH:mm');
    const endTime = moment.tz(utcTime.EndTime, 'HH:mm', 'GMT').tz(this.state.timezone).format('HH:mm');

    const [startHour, startMin] = startTime.split(':');
    const [endHour, endMin] = endTime.split(':');

    const startTimeUser = new Date();
    const endTimeUser = new Date();

    startTimeUser.setHours(startHour);
    startTimeUser.setMinutes(startMin);

    endTimeUser.setHours(endHour);
    endTimeUser.setMinutes(endMin);

    return {
      startTime: startTimeUser,
      endTime: endTimeUser,
    };
  }

  customToTimeObject() {
    const startTime = this.state.timeWindow.StartTime;
    const endTime = this.state.timeWindow.EndTime;

    const [startHour, startMin] = startTime.split(':');
    const [endHour, endMin] = endTime.split(':');

    const startTimeObject = new Date();
    const endTimeObject = new Date();

    startTimeObject.setHours(startHour);
    startTimeObject.setMinutes(startMin);

    endTimeObject.setHours(endHour);
    endTimeObject.setMinutes(endMin);

    this.state.startTimeUtc = startTimeObject;
    this.state.endTimeUtc = endTimeObject;
  }

  loadUserTime(utcTime) {
    const { startTime, endTime } = this.utcToTime(utcTime);

    this.state.startTimeUser = startTime;
    this.state.endTimeUser = endTime;
  }

  $onInit() {
    const currentUserTimezone = moment.tz.guess();

    return this.$async(async () => {
      const endpointId = this.EndpointProvider.endpointID();
      const endpoint = await this.EndpointService.endpoint(endpointId);

      this.state = {
        timeWindow: endpoint.ChangeWindow,
        startTimeUser: '',
        endTimeUser: '',
        startTimeUtc: '',
        endTimeUtc: '',
        timezone: currentUserTimezone,
        dst: moment().isDST(),
      };

      this.loadUserTime(this.state.timeWindow);
      this.customToTimeObject();
    });
  }
}
