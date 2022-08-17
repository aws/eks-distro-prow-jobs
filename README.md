# EKS Distro Prow Jobs

This repository contains Prow Job configuration for the EKS installation of
Prow, which is available at https://prow.eks.amazonaws.com/.

For more info on how to write a prow job, read the [test-infra
introduction](https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md)
to prow jobs.

## Creating a New EKS Distro Prow Job
All EKS Distro Prow jobs are automatically generated from a template.
In order to create a new prow job, add a template containing your job details to the relevant folder in the [`templates/jobs`](https://github.com/aws/eks-distro-prow-jobs/tree/main/templater/jobs) directory (`presubmits` for presubmits, etc), 
run [`make all`](https://github.com/aws/eks-distro-prow-jobs/blob/main/templater/Makefile#L4) in the [`templater` Makefile](https://github.com/aws/eks-distro-prow-jobs/blob/main/templater/Makefile) to generate job specs from the template, and commit the generated job specs output by the templater.

Manually creating the job specs without the accompanying template will result in a presubmit verification failure.

## Contributing

Please read our [CONTRIBUTING](CONTRIBUTING.md) guide before making a pull
request.

Refer to the [documentation](docs/prowjobs.md) for information on how to add or update Prowjobs.

## Security

If you discover a potential security issue in this project, or think you may
have discovered a security issue, we ask that you notify AWS Security via our
[vulnerability reporting
page](http://aws.amazon.com/security/vulnerability-reporting/). Please do
**not** create a public GitHub issue.

## License

This project is licensed under the Apache-2.0 License.
